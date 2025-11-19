package config

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gymshark/go-easy-config/loader/generic"
	"github.com/gymshark/go-easy-config/utils"
)

type TestConfig struct {
	EnvVar1    string `env:"TEST_ENV_VAR1"`
	SecretVar1 string
	CmdVar1    string `clap:"--cmdvar1"`
}

const (
	CmdValue = "cmd_value"
	EnvValue = "env_value"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{SecretVar1: "secret_value"}
	config := NewConfigHandler[TestConfig]()
	if err := config.Load(cfg); err != nil {
		t.Fatalf("ConfigHandler.Load failed: %v", err)
	}
	if cfg.EnvVar1 != EnvValue {
		t.Errorf("EnvVar1 not loaded, got: %s", cfg.EnvVar1)
	}
}

func TestValidateConfig(t *testing.T) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{EnvVar1: EnvValue}
	config := NewConfigHandler[TestConfig]()
	if err := config.Validate(cfg); err != nil {
		t.Fatalf("ConfigHandler.Validate failed: %v", err)
	}
}

func TestLoadAndValidateConfig(t *testing.T) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{SecretVar1: "secret_value"}
	config := NewConfigHandler[TestConfig]()
	if err := config.LoadAndValidate(cfg); err != nil {
		t.Fatalf("ConfigHandler.LoadAndValidate failed: %v", err)
	}
	if cfg.EnvVar1 != EnvValue {
		t.Errorf("EnvVar1 not loaded, got: %s", cfg.EnvVar1)
	}
}

func TestWithValidator(t *testing.T) {
	customValidator := DefaultConfigValidator()
	handler := NewConfigHandler[TestConfig](WithValidator[TestConfig](customValidator))
	if handler.Validator != customValidator {
		t.Errorf("WithValidator did not set custom validator")
	}
}

func TestWithLoaders(t *testing.T) {
	customLoader := &generic.EnvironmentLoader[TestConfig]{}
	handler := NewConfigHandler[TestConfig](WithLoaders(customLoader))
	if len(handler.Loaders) != 1 {
		t.Errorf("WithLoaders did not set custom loaders")
	}
	if _, ok := handler.Loaders[0].(*generic.EnvironmentLoader[TestConfig]); !ok {
		t.Errorf("WithLoaders did not set correct loader type")
	}
}

func TestIsZero_AllTypes(t *testing.T) {
	var (
		str                    = ""
		b                      = false
		intVal     int64       = 0
		uintVal    uint64      = 0
		floatVal   float64     = 0
		complexVal complex128  = 0
		ptr        *int        = nil
		iface      interface{} = nil
		arr        [2]int
		ch         chan int
		m          map[string]int
		slice      []int
	)
	cases := []struct {
		val  reflect.Value
		want bool
	}{
		{reflect.ValueOf(str), true},
		{reflect.ValueOf(b), true},
		{reflect.ValueOf(intVal), true},
		{reflect.ValueOf(uintVal), true},
		{reflect.ValueOf(floatVal), true},
		{reflect.ValueOf(complexVal), true},
		{reflect.ValueOf(ptr), true},
		{reflect.ValueOf(iface), false}, // interface is nil, expect false due to Go reflect semantics
		{reflect.ValueOf(arr), true},    // array of zero values, expect true
		{reflect.ValueOf(ch), true},
		{reflect.ValueOf(m), true},
		{reflect.ValueOf(slice), true},
	}
	for i, c := range cases {
		if got := utils.IsZero(c.val); got != c.want {
			t.Errorf("isZero case %d failed: got %v, want %v", i, got, c.want)
		}
	}
}

func BenchmarkEnvironmentLoader_Load(b *testing.B) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{}
	loader := &generic.EnvironmentLoader[TestConfig]{}
	for i := 0; i < b.N; i++ {
		_ = loader.Load(cfg)
	}
}

func BenchmarkCommandLineLoader_Load(b *testing.B) {
	cfg := &TestConfig{}
	args := []string{"--cmdvar1", CmdValue}
	loader := &generic.CommandLineLoader[TestConfig]{Args: args}
	for i := 0; i < b.N; i++ {
		_ = loader.Load(cfg)
	}
}

func BenchmarkInterpolatingChainLoader_Load_NoInterpolation(b *testing.B) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{}
	loaders := []Loader[TestConfig]{
		&generic.EnvironmentLoader[TestConfig]{},
		&generic.CommandLineLoader[TestConfig]{Args: []string{"--cmdvar1", CmdValue}},
	}
	chain := &InterpolatingChainLoader[TestConfig]{Loaders: loaders}
	for i := 0; i < b.N; i++ {
		_ = chain.Load(cfg)
	}
}

func BenchmarkInterpolatingChainLoader_Load_WithShortCircuit(b *testing.B) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{}
	loaders := []Loader[TestConfig]{
		&generic.EnvironmentLoader[TestConfig]{},
		&generic.CommandLineLoader[TestConfig]{Args: []string{"--cmdvar1", CmdValue}},
	}
	chain := &InterpolatingChainLoader[TestConfig]{
		Loaders:      loaders,
		ShortCircuit: true,
	}
	for i := 0; i < b.N; i++ {
		_ = chain.Load(cfg)
	}
}

func BenchmarkConfigHandler_Load(b *testing.B) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{}
	config := NewConfigHandler[TestConfig]()
	for i := 0; i < b.N; i++ {
		_ = config.Load(cfg)
	}
}

func BenchmarkConfigHandler_LoadAndValidate(b *testing.B) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{}
	config := NewConfigHandler[TestConfig]()
	for i := 0; i < b.N; i++ {
		_ = config.LoadAndValidate(cfg)
	}
}

// Test configuration structs for validation error tests
type ValidationTestConfig struct {
	RequiredField string `validate:"required"`
	MinField      int    `validate:"min=10"`
	EmailField    string `validate:"email"`
}

type MultipleValidationErrorsConfig struct {
	Field1 string `validate:"required"`
	Field2 int    `validate:"min=5"`
	Field3 string `validate:"email"`
}

// TestValidationError_ReturnedOnFailure tests that ValidationError is returned when validation fails
func TestValidationError_ReturnedOnFailure(t *testing.T) {
	handler := NewConfigHandler[ValidationTestConfig]()
	cfg := &ValidationTestConfig{
		RequiredField: "", // Missing required field
		MinField:      5,  // Below minimum
		EmailField:    "invalid-email",
	}

	err := handler.Validate(cfg)
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	// Check that error is ValidationError
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError, got %T: %v", err, err)
	}

	// Verify ValidationError fields
	if validationErr.FieldName != "<multiple>" {
		t.Errorf("Expected FieldName '<multiple>', got '%s'", validationErr.FieldName)
	}
	if validationErr.Rule != "<multiple>" {
		t.Errorf("Expected Rule '<multiple>', got '%s'", validationErr.Rule)
	}
	if validationErr.Err == nil {
		t.Error("Expected underlying error to be set")
	}
}

// TestValidationError_UnderlyingErrorAccessible tests that underlying validator errors are accessible via errors.As
func TestValidationError_UnderlyingErrorAccessible(t *testing.T) {
	handler := NewConfigHandler[ValidationTestConfig]()
	cfg := &ValidationTestConfig{
		RequiredField: "", // Missing required field
	}

	err := handler.Validate(cfg)
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	// Check that we can extract ValidationError
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatal("Could not extract ValidationError using errors.As")
	}

	// Check that underlying validator error is accessible
	if validationErr.Err == nil {
		t.Fatal("Expected underlying validator error to be set")
	}

	// Verify we can access the validator.ValidationErrors type
	var validatorErrs validator.ValidationErrors
	if !errors.As(validationErr.Err, &validatorErrs) {
		t.Errorf("Could not extract validator.ValidationErrors from underlying error")
	}
}

// TestValidationError_MessageIncludesContext tests that error messages include field and rule context
func TestValidationError_MessageIncludesContext(t *testing.T) {
	handler := NewConfigHandler[ValidationTestConfig]()
	cfg := &ValidationTestConfig{
		RequiredField: "", // Missing required field
	}

	err := handler.Validate(cfg)
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	// Check error message format
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message is empty")
	}

	// Verify message includes field and rule context
	expectedSubstrings := []string{"validation failed", "field", "<multiple>", "rule"}
	for _, substr := range expectedSubstrings {
		if !contains(errMsg, substr) {
			t.Errorf("Error message missing expected substring '%s': %s", substr, errMsg)
		}
	}
}

// TestValidationError_MultipleFieldFailures tests ValidationError with multiple field failures
func TestValidationError_MultipleFieldFailures(t *testing.T) {
	handler := NewConfigHandler[MultipleValidationErrorsConfig]()
	cfg := &MultipleValidationErrorsConfig{
		Field1: "",             // Missing required
		Field2: 2,              // Below minimum
		Field3: "not-an-email", // Invalid email
	}

	err := handler.Validate(cfg)
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	// Verify the underlying validator error contains multiple failures
	var validatorErrs validator.ValidationErrors
	if !errors.As(validationErr.Err, &validatorErrs) {
		t.Fatal("Could not extract validator.ValidationErrors")
	}

	// Should have 3 validation errors
	if len(validatorErrs) != 3 {
		t.Errorf("Expected 3 validation errors, got %d", len(validatorErrs))
	}
}

// TestValidationError_SuccessfulValidation tests that no error is returned for valid configuration
func TestValidationError_SuccessfulValidation(t *testing.T) {
	handler := NewConfigHandler[ValidationTestConfig]()
	cfg := &ValidationTestConfig{
		RequiredField: "present",
		MinField:      15,
		EmailField:    "valid@example.com",
	}

	err := handler.Validate(cfg)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}

// TestLoadAndValidate_ReturnsValidationError tests that LoadAndValidate returns ValidationError on validation failure
func TestLoadAndValidate_ReturnsValidationError(t *testing.T) {
	// Set up environment for loading
	os.Setenv("REQUIRED_FIELD", "loaded")
	defer os.Unsetenv("REQUIRED_FIELD")

	type LoadValidateConfig struct {
		RequiredField string `env:"REQUIRED_FIELD" validate:"required"`
		MinField      int    `validate:"min=10"` // Will fail validation (zero value)
	}

	handler := NewConfigHandler[LoadValidateConfig]()
	cfg := &LoadValidateConfig{}

	err := handler.LoadAndValidate(cfg)
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	// Verify it's a ValidationError
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError from LoadAndValidate, got %T", err)
	}

	// Verify the field was loaded before validation failed
	if cfg.RequiredField != "loaded" {
		t.Errorf("Expected RequiredField to be loaded before validation, got '%s'", cfg.RequiredField)
	}
}

// Mock loader that always fails with LoaderError for testing
type FailingLoader[T any] struct {
	ErrorToReturn error
}

func (f *FailingLoader[T]) Load(cfg *T) error {
	if f.ErrorToReturn != nil {
		return f.ErrorToReturn
	}
	return &LoaderError{
		LoaderType: "FailingLoader",
		Operation:  "mock load",
		Source:     "test",
		Err:        errors.New("mock loader failure"),
	}
}

// TestHandler_Load_PassesThroughLoaderError tests that Load method passes through loader errors without wrapping
func TestHandler_Load_PassesThroughLoaderError(t *testing.T) {
	// Create a loader that returns a LoaderError
	expectedErr := &LoaderError{
		LoaderType: "TestLoader",
		Operation:  "test operation",
		Source:     "test.json",
		Err:        errors.New("test error"),
	}

	failingLoader := &FailingLoader[TestConfig]{ErrorToReturn: expectedErr}
	handler := NewConfigHandler[TestConfig](WithLoaders[TestConfig](failingLoader))

	cfg := &TestConfig{}
	err := handler.Load(cfg)

	if err == nil {
		t.Fatal("Expected error from Load, got nil")
	}

	// Verify the error is the exact same LoaderError (not wrapped)
	var loaderErr *LoaderError
	if !errors.As(err, &loaderErr) {
		t.Fatalf("Expected LoaderError, got %T: %v", err, err)
	}

	// Verify it's the same error instance
	if loaderErr != expectedErr {
		t.Error("Expected Load to return the exact same LoaderError instance without wrapping")
	}

	// Verify error details are preserved
	if loaderErr.LoaderType != "TestLoader" {
		t.Errorf("Expected LoaderType 'TestLoader', got '%s'", loaderErr.LoaderType)
	}
	if loaderErr.Operation != "test operation" {
		t.Errorf("Expected Operation 'test operation', got '%s'", loaderErr.Operation)
	}
	if loaderErr.Source != "test.json" {
		t.Errorf("Expected Source 'test.json', got '%s'", loaderErr.Source)
	}
}

// TestHandler_Validate_ReturnsValidationError tests that Validate method returns ValidationError
func TestHandler_Validate_ReturnsValidationError(t *testing.T) {
	handler := NewConfigHandler[ValidationTestConfig]()
	cfg := &ValidationTestConfig{
		RequiredField: "", // Missing required field
		MinField:      5,  // Below minimum
	}

	err := handler.Validate(cfg)
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	// Verify it's a ValidationError
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError, got %T: %v", err, err)
	}

	// Verify ValidationError fields
	if validationErr.FieldName != "<multiple>" {
		t.Errorf("Expected FieldName '<multiple>', got '%s'", validationErr.FieldName)
	}
	if validationErr.Rule != "<multiple>" {
		t.Errorf("Expected Rule '<multiple>', got '%s'", validationErr.Rule)
	}
	if validationErr.Err == nil {
		t.Error("Expected underlying error to be set")
	}
}

// TestHandler_LoadAndValidate_ReturnsLoaderErrorOnLoadFailure tests that LoadAndValidate returns LoaderError when loading fails
func TestHandler_LoadAndValidate_ReturnsLoaderErrorOnLoadFailure(t *testing.T) {
	// Create a loader that fails
	expectedErr := &LoaderError{
		LoaderType: "FailingLoader",
		Operation:  "load config",
		Source:     "config.json",
		Err:        errors.New("file not found"),
	}

	failingLoader := &FailingLoader[ValidationTestConfig]{ErrorToReturn: expectedErr}
	handler := NewConfigHandler[ValidationTestConfig](WithLoaders[ValidationTestConfig](failingLoader))

	cfg := &ValidationTestConfig{}
	err := handler.LoadAndValidate(cfg)

	if err == nil {
		t.Fatal("Expected error from LoadAndValidate, got nil")
	}

	// Verify it's a LoaderError (not ValidationError)
	var loaderErr *LoaderError
	if !errors.As(err, &loaderErr) {
		t.Fatalf("Expected LoaderError from LoadAndValidate when loading fails, got %T: %v", err, err)
	}

	// Verify it's the same error
	if loaderErr != expectedErr {
		t.Error("Expected LoadAndValidate to return the exact same LoaderError instance")
	}

	// Verify it's NOT a ValidationError
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		t.Error("LoadAndValidate should not return ValidationError when loading fails")
	}
}

// TestHandler_LoadAndValidate_ReturnsValidationErrorOnValidationFailure tests that LoadAndValidate returns ValidationError when validation fails
func TestHandler_LoadAndValidate_ReturnsValidationErrorOnValidationFailure(t *testing.T) {
	// Use a loader that succeeds but produces invalid config
	handler := NewConfigHandler[ValidationTestConfig]()
	cfg := &ValidationTestConfig{
		RequiredField: "", // Missing required field - will fail validation
		MinField:      5,  // Below minimum - will fail validation
	}

	err := handler.LoadAndValidate(cfg)

	if err == nil {
		t.Fatal("Expected validation error from LoadAndValidate, got nil")
	}

	// Verify it's a ValidationError (not LoaderError)
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError from LoadAndValidate when validation fails, got %T: %v", err, err)
	}

	// Verify ValidationError fields
	if validationErr.FieldName != "<multiple>" {
		t.Errorf("Expected FieldName '<multiple>', got '%s'", validationErr.FieldName)
	}
	if validationErr.Rule != "<multiple>" {
		t.Errorf("Expected Rule '<multiple>', got '%s'", validationErr.Rule)
	}

	// Verify it's NOT a LoaderError
	var loaderErr *LoaderError
	if errors.As(err, &loaderErr) {
		t.Error("LoadAndValidate should not return LoaderError when validation fails")
	}
}

// TestHandler_ErrorsAs_WorksCorrectly tests that errors.As works correctly for all error types from handler methods
func TestHandler_ErrorsAs_WorksCorrectly(t *testing.T) {
	t.Run("Load returns LoaderError extractable with errors.As", func(t *testing.T) {
		expectedErr := &LoaderError{
			LoaderType: "TestLoader",
			Operation:  "test",
			Err:        errors.New("test error"),
		}

		failingLoader := &FailingLoader[TestConfig]{ErrorToReturn: expectedErr}
		handler := NewConfigHandler[TestConfig](WithLoaders[TestConfig](failingLoader))

		cfg := &TestConfig{}
		err := handler.Load(cfg)

		var loaderErr *LoaderError
		if !errors.As(err, &loaderErr) {
			t.Fatal("errors.As failed to extract LoaderError from Load method")
		}

		if loaderErr.LoaderType != "TestLoader" {
			t.Errorf("Expected LoaderType 'TestLoader', got '%s'", loaderErr.LoaderType)
		}
	})

	t.Run("Validate returns ValidationError extractable with errors.As", func(t *testing.T) {
		handler := NewConfigHandler[ValidationTestConfig]()
		cfg := &ValidationTestConfig{
			RequiredField: "", // Invalid
		}

		err := handler.Validate(cfg)

		var validationErr *ValidationError
		if !errors.As(err, &validationErr) {
			t.Fatal("errors.As failed to extract ValidationError from Validate method")
		}

		if validationErr.FieldName != "<multiple>" {
			t.Errorf("Expected FieldName '<multiple>', got '%s'", validationErr.FieldName)
		}
	})

	t.Run("LoadAndValidate returns LoaderError extractable with errors.As on load failure", func(t *testing.T) {
		expectedErr := &LoaderError{
			LoaderType: "TestLoader",
			Operation:  "test",
			Err:        errors.New("test error"),
		}

		failingLoader := &FailingLoader[ValidationTestConfig]{ErrorToReturn: expectedErr}
		handler := NewConfigHandler[ValidationTestConfig](WithLoaders[ValidationTestConfig](failingLoader))

		cfg := &ValidationTestConfig{}
		err := handler.LoadAndValidate(cfg)

		var loaderErr *LoaderError
		if !errors.As(err, &loaderErr) {
			t.Fatal("errors.As failed to extract LoaderError from LoadAndValidate on load failure")
		}

		if loaderErr.LoaderType != "TestLoader" {
			t.Errorf("Expected LoaderType 'TestLoader', got '%s'", loaderErr.LoaderType)
		}
	})

	t.Run("LoadAndValidate returns ValidationError extractable with errors.As on validation failure", func(t *testing.T) {
		handler := NewConfigHandler[ValidationTestConfig]()
		cfg := &ValidationTestConfig{
			RequiredField: "", // Invalid
		}

		err := handler.LoadAndValidate(cfg)

		var validationErr *ValidationError
		if !errors.As(err, &validationErr) {
			t.Fatal("errors.As failed to extract ValidationError from LoadAndValidate on validation failure")
		}

		if validationErr.FieldName != "<multiple>" {
			t.Errorf("Expected FieldName '<multiple>', got '%s'", validationErr.FieldName)
		}
	})

	t.Run("Underlying validator errors accessible through ValidationError", func(t *testing.T) {
		handler := NewConfigHandler[ValidationTestConfig]()
		cfg := &ValidationTestConfig{
			RequiredField: "", // Invalid
		}

		err := handler.Validate(cfg)

		// Extract ValidationError
		var validationErr *ValidationError
		if !errors.As(err, &validationErr) {
			t.Fatal("Failed to extract ValidationError")
		}

		// Extract underlying validator.ValidationErrors
		var validatorErrs validator.ValidationErrors
		if !errors.As(validationErr.Err, &validatorErrs) {
			t.Fatal("errors.As failed to extract validator.ValidationErrors from ValidationError.Err")
		}

		// Verify we got the validator errors
		if len(validatorErrs) == 0 {
			t.Error("Expected validator errors to be present")
		}
	})
}
