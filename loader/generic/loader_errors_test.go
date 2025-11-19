package generic

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/gymshark/go-easy-config/loader"
)

// TestEnvironmentLoader_ReturnsLoaderError tests that EnvironmentLoader returns LoaderError on failures
func TestEnvironmentLoader_ReturnsLoaderError(t *testing.T) {
	type invalidConfig struct {
		Port string `env:"PORT,required"`
	}

	// Clear PORT to trigger required field error
	os.Unsetenv("PORT")

	cfg := &invalidConfig{}
	envLoader := &EnvironmentLoader[invalidConfig]{}
	err := envLoader.Load(cfg)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Test that error is LoaderError
	var loaderErr *loader.LoaderError
	if !errors.As(err, &loaderErr) {
		t.Fatalf("expected LoaderError, got %T: %v", err, err)
	}

	// Test LoaderError fields
	if loaderErr.LoaderType != "EnvironmentLoader" {
		t.Errorf("expected LoaderType 'EnvironmentLoader', got '%s'", loaderErr.LoaderType)
	}

	if loaderErr.Operation != "parse environment variables" {
		t.Errorf("expected Operation 'parse environment variables', got '%s'", loaderErr.Operation)
	}

	// Test that underlying error is accessible
	if loaderErr.Err == nil {
		t.Error("expected underlying error to be set")
	}

	// Test error message format
	errMsg := err.Error()
	if !strings.Contains(errMsg, "EnvironmentLoader") {
		t.Errorf("error message should contain 'EnvironmentLoader', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "parse environment variables") {
		t.Errorf("error message should contain 'parse environment variables', got: %s", errMsg)
	}
}

// TestCommandLineLoader_ReturnsLoaderError tests that CommandLineLoader returns LoaderError on failures
func TestCommandLineLoader_ReturnsLoaderError(t *testing.T) {
	type testConfig struct {
		Port int `clap:"--port"`
	}

	// Invalid argument format (non-integer for int field)
	cfg := &testConfig{}
	cmdLoader := &CommandLineLoader[testConfig]{Args: []string{"--port", "invalid"}}
	err := cmdLoader.Load(cfg)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Test that error is LoaderError
	var loaderErr *loader.LoaderError
	if !errors.As(err, &loaderErr) {
		t.Fatalf("expected LoaderError, got %T: %v", err, err)
	}

	// Test LoaderError fields
	if loaderErr.LoaderType != "CommandLineLoader" {
		t.Errorf("expected LoaderType 'CommandLineLoader', got '%s'", loaderErr.LoaderType)
	}

	if loaderErr.Operation != "parse command line arguments" {
		t.Errorf("expected Operation 'parse command line arguments', got '%s'", loaderErr.Operation)
	}

	// Test that underlying error is accessible
	if loaderErr.Err == nil {
		t.Error("expected underlying error to be set")
	}

	// Test error message format
	errMsg := err.Error()
	if !strings.Contains(errMsg, "CommandLineLoader") {
		t.Errorf("error message should contain 'CommandLineLoader', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "parse command line arguments") {
		t.Errorf("error message should contain 'parse command line arguments', got: %s", errMsg)
	}
}

// TestJSONLoader_ReturnsLoaderError_FileNotFound tests JSONLoader returns LoaderError for missing files
func TestJSONLoader_ReturnsLoaderError_FileNotFound(t *testing.T) {
	type testConfig struct {
		Field string `json:"field"`
	}

	cfg := &testConfig{}
	jsonLoader := &JSONLoader[testConfig]{Source: "nonexistent.json"}
	err := jsonLoader.Load(cfg)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Test that error is LoaderError
	var loaderErr *loader.LoaderError
	if !errors.As(err, &loaderErr) {
		t.Fatalf("expected LoaderError, got %T: %v", err, err)
	}

	// Test LoaderError fields
	if loaderErr.LoaderType != "JSONLoader" {
		t.Errorf("expected LoaderType 'JSONLoader', got '%s'", loaderErr.LoaderType)
	}

	if loaderErr.Operation != "read file" {
		t.Errorf("expected Operation 'read file', got '%s'", loaderErr.Operation)
	}

	if loaderErr.Source != "nonexistent.json" {
		t.Errorf("expected Source 'nonexistent.json', got '%s'", loaderErr.Source)
	}

	// Test that underlying error is accessible
	if loaderErr.Err == nil {
		t.Error("expected underlying error to be set")
	}

	// Test error message format includes source
	errMsg := err.Error()
	if !strings.Contains(errMsg, "JSONLoader") {
		t.Errorf("error message should contain 'JSONLoader', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "read file") {
		t.Errorf("error message should contain 'read file', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "nonexistent.json") {
		t.Errorf("error message should contain source 'nonexistent.json', got: %s", errMsg)
	}
}

// TestJSONLoader_ReturnsLoaderError_InvalidJSON tests JSONLoader returns LoaderError for invalid JSON
func TestJSONLoader_ReturnsLoaderError_InvalidJSON(t *testing.T) {
	type testConfig struct {
		Field string `json:"field"`
	}

	path := "invalid_test.json"
	if err := writeTestJSONFile(path, "not valid json"); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	defer os.Remove(path)

	cfg := &testConfig{}
	jsonLoader := &JSONLoader[testConfig]{Source: path}
	err := jsonLoader.Load(cfg)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Test that error is LoaderError
	var loaderErr *loader.LoaderError
	if !errors.As(err, &loaderErr) {
		t.Fatalf("expected LoaderError, got %T: %v", err, err)
	}

	// Test LoaderError fields
	if loaderErr.LoaderType != "JSONLoader" {
		t.Errorf("expected LoaderType 'JSONLoader', got '%s'", loaderErr.LoaderType)
	}

	if loaderErr.Operation != "unmarshal JSON" {
		t.Errorf("expected Operation 'unmarshal JSON', got '%s'", loaderErr.Operation)
	}

	if loaderErr.Source != path {
		t.Errorf("expected Source '%s', got '%s'", path, loaderErr.Source)
	}

	// Test error message format
	errMsg := err.Error()
	if !strings.Contains(errMsg, "JSONLoader") {
		t.Errorf("error message should contain 'JSONLoader', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "unmarshal JSON") {
		t.Errorf("error message should contain 'unmarshal JSON', got: %s", errMsg)
	}
}

// TestJSONLoader_ReturnsLoaderError_UnsupportedSourceType tests JSONLoader returns LoaderError for invalid source types
func TestJSONLoader_ReturnsLoaderError_UnsupportedSourceType(t *testing.T) {
	type testConfig struct {
		Field string `json:"field"`
	}

	cfg := &testConfig{}
	jsonLoader := &JSONLoader[testConfig]{Source: 12345} // Invalid source type
	err := jsonLoader.Load(cfg)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Test that error is LoaderError
	var loaderErr *loader.LoaderError
	if !errors.As(err, &loaderErr) {
		t.Fatalf("expected LoaderError, got %T: %v", err, err)
	}

	// Test LoaderError fields
	if loaderErr.LoaderType != "JSONLoader" {
		t.Errorf("expected LoaderType 'JSONLoader', got '%s'", loaderErr.LoaderType)
	}

	if loaderErr.Operation != "validate source type" {
		t.Errorf("expected Operation 'validate source type', got '%s'", loaderErr.Operation)
	}

	if !strings.Contains(loaderErr.Source, "int") {
		t.Errorf("expected Source to contain type 'int', got '%s'", loaderErr.Source)
	}
}

// TestYAMLLoader_ReturnsLoaderError_FileNotFound tests YAMLLoader returns LoaderError for missing files
func TestYAMLLoader_ReturnsLoaderError_FileNotFound(t *testing.T) {
	type testConfig struct {
		Field string `yaml:"field"`
	}

	cfg := &testConfig{}
	yamlLoader := &YAMLLoader[testConfig]{Source: "nonexistent.yaml"}
	err := yamlLoader.Load(cfg)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Test that error is LoaderError
	var loaderErr *loader.LoaderError
	if !errors.As(err, &loaderErr) {
		t.Fatalf("expected LoaderError, got %T: %v", err, err)
	}

	// Test LoaderError fields
	if loaderErr.LoaderType != "YAMLLoader" {
		t.Errorf("expected LoaderType 'YAMLLoader', got '%s'", loaderErr.LoaderType)
	}

	if loaderErr.Operation != "read file" {
		t.Errorf("expected Operation 'read file', got '%s'", loaderErr.Operation)
	}

	if loaderErr.Source != "nonexistent.yaml" {
		t.Errorf("expected Source 'nonexistent.yaml', got '%s'", loaderErr.Source)
	}

	// Test error message format includes source
	errMsg := err.Error()
	if !strings.Contains(errMsg, "YAMLLoader") {
		t.Errorf("error message should contain 'YAMLLoader', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "nonexistent.yaml") {
		t.Errorf("error message should contain source 'nonexistent.yaml', got: %s", errMsg)
	}
}

// TestINILoader_ReturnsLoaderError_FileNotFound tests INILoader returns LoaderError for missing files
func TestINILoader_ReturnsLoaderError_FileNotFound(t *testing.T) {
	type testConfig struct {
		Field string `ini:"field"`
	}

	cfg := &testConfig{}
	iniLoader := &IniLoader[testConfig]{Source: "nonexistent.ini"}
	err := iniLoader.Load(cfg)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Test that error is LoaderError
	var loaderErr *loader.LoaderError
	if !errors.As(err, &loaderErr) {
		t.Fatalf("expected LoaderError, got %T: %v", err, err)
	}

	// Test LoaderError fields
	if loaderErr.LoaderType != "INILoader" {
		t.Errorf("expected LoaderType 'INILoader', got '%s'", loaderErr.LoaderType)
	}

	if loaderErr.Operation != "load INI file" {
		t.Errorf("expected Operation 'load INI file', got '%s'", loaderErr.Operation)
	}

	if loaderErr.Source != "nonexistent.ini" {
		t.Errorf("expected Source 'nonexistent.ini', got '%s'", loaderErr.Source)
	}

	// Test error message format includes source
	errMsg := err.Error()
	if !strings.Contains(errMsg, "INILoader") {
		t.Errorf("error message should contain 'INILoader', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "nonexistent.ini") {
		t.Errorf("error message should contain source 'nonexistent.ini', got: %s", errMsg)
	}
}

// TestLoaderError_ErrorChainTraversal tests that wrapped errors are accessible via errors.As
func TestLoaderError_ErrorChainTraversal(t *testing.T) {
	type testConfig struct {
		Field string `json:"field"`
	}

	cfg := &testConfig{}
	jsonLoader := &JSONLoader[testConfig]{Source: "nonexistent.json"}
	err := jsonLoader.Load(cfg)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Test that we can extract LoaderError
	var loaderErr *loader.LoaderError
	if !errors.As(err, &loaderErr) {
		t.Fatal("failed to extract LoaderError with errors.As")
	}

	// Test that we can access the underlying error
	underlyingErr := errors.Unwrap(err)
	if underlyingErr == nil {
		t.Fatal("expected underlying error, got nil")
	}

	// Verify the underlying error is the same as loaderErr.Err
	if underlyingErr != loaderErr.Err {
		t.Error("Unwrap() should return the same error as loaderErr.Err")
	}
}

// TestAllLoaders_ErrorMessageConsistency tests that all loaders follow consistent error message format
func TestAllLoaders_ErrorMessageConsistency(t *testing.T) {
	tests := []struct {
		name        string
		loaderType  string
		operation   string
		source      string
		setupLoader func() error
	}{
		{
			name:       "JSONLoader file not found",
			loaderType: "JSONLoader",
			operation:  "read file",
			source:     "test.json",
			setupLoader: func() error {
				type cfg struct {
					F string `json:"f"`
				}
				return (&JSONLoader[cfg]{Source: "test.json"}).Load(&cfg{})
			},
		},
		{
			name:       "YAMLLoader file not found",
			loaderType: "YAMLLoader",
			operation:  "read file",
			source:     "test.yaml",
			setupLoader: func() error {
				type cfg struct {
					F string `yaml:"f"`
				}
				return (&YAMLLoader[cfg]{Source: "test.yaml"}).Load(&cfg{})
			},
		},
		{
			name:       "INILoader file not found",
			loaderType: "INILoader",
			operation:  "load INI file",
			source:     "test.ini",
			setupLoader: func() error {
				type cfg struct {
					F string `ini:"f"`
				}
				return (&IniLoader[cfg]{Source: "test.ini"}).Load(&cfg{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setupLoader()
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var loaderErr *loader.LoaderError
			if !errors.As(err, &loaderErr) {
				t.Fatalf("expected LoaderError, got %T", err)
			}

			// Verify consistent fields
			if loaderErr.LoaderType != tt.loaderType {
				t.Errorf("expected LoaderType '%s', got '%s'", tt.loaderType, loaderErr.LoaderType)
			}

			if loaderErr.Operation != tt.operation {
				t.Errorf("expected Operation '%s', got '%s'", tt.operation, loaderErr.Operation)
			}

			if loaderErr.Source != tt.source {
				t.Errorf("expected Source '%s', got '%s'", tt.source, loaderErr.Source)
			}

			// Verify error message format
			errMsg := err.Error()
			expectedParts := []string{tt.loaderType, tt.operation, tt.source}
			for _, part := range expectedParts {
				if !strings.Contains(errMsg, part) {
					t.Errorf("error message should contain '%s', got: %s", part, errMsg)
				}
			}
		})
	}
}
