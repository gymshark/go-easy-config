package config

import (
	"errors"
	"fmt"
	"testing"
)

// TestLoaderError_Error tests the Error() method formatting for LoaderError
func TestLoaderError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *LoaderError
		expected string
	}{
		{
			name: "with source",
			err: &LoaderError{
				LoaderType: "JSONLoader",
				Operation:  "read file",
				Source:     "config.json",
				Err:        fmt.Errorf("file not found"),
			},
			expected: "JSONLoader error during read file (source: config.json): file not found",
		},
		{
			name: "without source",
			err: &LoaderError{
				LoaderType: "EnvironmentLoader",
				Operation:  "parse environment variables",
				Err:        fmt.Errorf("invalid format"),
			},
			expected: "EnvironmentLoader error during parse environment variables: invalid format",
		},
		{
			name: "with empty source",
			err: &LoaderError{
				LoaderType: "YAMLLoader",
				Operation:  "unmarshal",
				Source:     "",
				Err:        fmt.Errorf("yaml syntax error"),
			},
			expected: "YAMLLoader error during unmarshal: yaml syntax error",
		},
		{
			name: "AWS loader with source",
			err: &LoaderError{
				LoaderType: "SecretsManagerLoader",
				Operation:  "fetch secrets",
				Source:     "/prod/api/key",
				Err:        fmt.Errorf("access denied"),
			},
			expected: "SecretsManagerLoader error during fetch secrets (source: /prod/api/key): access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestLoaderError_Unwrap tests that Unwrap returns the wrapped error correctly
func TestLoaderError_Unwrap(t *testing.T) {
	underlying := fmt.Errorf("underlying error")
	err := &LoaderError{
		LoaderType: "JSONLoader",
		Operation:  "unmarshal",
		Err:        underlying,
	}

	unwrapped := errors.Unwrap(err)
	if unwrapped != underlying {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, underlying)
	}
}

// TestLoaderError_As tests that errors.As can extract LoaderError from error chains
func TestLoaderError_As(t *testing.T) {
	loaderErr := &LoaderError{
		LoaderType: "YAMLLoader",
		Operation:  "read file",
		Source:     "config.yaml",
		Err:        fmt.Errorf("test error"),
	}

	// Test direct error
	var extractedErr *LoaderError
	if !errors.As(loaderErr, &extractedErr) {
		t.Error("errors.As failed to extract LoaderError")
	}

	if extractedErr.LoaderType != "YAMLLoader" {
		t.Errorf("LoaderType = %v, want YAMLLoader", extractedErr.LoaderType)
	}
	if extractedErr.Operation != "read file" {
		t.Errorf("Operation = %v, want read file", extractedErr.Operation)
	}
	if extractedErr.Source != "config.yaml" {
		t.Errorf("Source = %v, want config.yaml", extractedErr.Source)
	}

	// Test wrapped error
	wrappedErr := fmt.Errorf("wrapper: %w", loaderErr)
	var extractedErr2 *LoaderError
	if !errors.As(wrappedErr, &extractedErr2) {
		t.Error("errors.As failed to extract LoaderError from wrapped error")
	}

	if extractedErr2.LoaderType != "YAMLLoader" {
		t.Errorf("LoaderType = %v, want YAMLLoader", extractedErr2.LoaderType)
	}
}

// TestValidationError_Error tests the Error() method formatting for ValidationError
func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "with value",
			err: &ValidationError{
				FieldName: "Port",
				Rule:      "min=1",
				Value:     "0",
				Err:       fmt.Errorf("validation failed"),
			},
			expected: "validation failed for field 'Port': rule 'min=1' failed (value: 0)",
		},
		{
			name: "without value",
			err: &ValidationError{
				FieldName: "Email",
				Rule:      "required",
				Err:       fmt.Errorf("validation failed"),
			},
			expected: "validation failed for field 'Email': rule 'required' failed",
		},
		{
			name: "with empty value",
			err: &ValidationError{
				FieldName: "Username",
				Rule:      "min=3",
				Value:     "",
				Err:       fmt.Errorf("validation failed"),
			},
			expected: "validation failed for field 'Username': rule 'min=3' failed",
		},
		{
			name: "multiple fields",
			err: &ValidationError{
				FieldName: "<multiple>",
				Rule:      "<multiple>",
				Err:       fmt.Errorf("multiple validation errors"),
			},
			expected: "validation failed for field '<multiple>': rule '<multiple>' failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestValidationError_Unwrap tests that Unwrap returns the wrapped error correctly
func TestValidationError_Unwrap(t *testing.T) {
	underlying := fmt.Errorf("underlying validation error")
	err := &ValidationError{
		FieldName: "Port",
		Rule:      "min=1",
		Err:       underlying,
	}

	unwrapped := errors.Unwrap(err)
	if unwrapped != underlying {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, underlying)
	}
}

// TestValidationError_As tests that errors.As can extract ValidationError from error chains
func TestValidationError_As(t *testing.T) {
	validationErr := &ValidationError{
		FieldName: "Port",
		Rule:      "min=1",
		Value:     "0",
		Err:       fmt.Errorf("test error"),
	}

	// Test direct error
	var extractedErr *ValidationError
	if !errors.As(validationErr, &extractedErr) {
		t.Error("errors.As failed to extract ValidationError")
	}

	if extractedErr.FieldName != "Port" {
		t.Errorf("FieldName = %v, want Port", extractedErr.FieldName)
	}
	if extractedErr.Rule != "min=1" {
		t.Errorf("Rule = %v, want min=1", extractedErr.Rule)
	}
	if extractedErr.Value != "0" {
		t.Errorf("Value = %v, want 0", extractedErr.Value)
	}

	// Test wrapped error
	wrappedErr := fmt.Errorf("wrapper: %w", validationErr)
	var extractedErr2 *ValidationError
	if !errors.As(wrappedErr, &extractedErr2) {
		t.Error("errors.As failed to extract ValidationError from wrapped error")
	}

	if extractedErr2.FieldName != "Port" {
		t.Errorf("FieldName = %v, want Port", extractedErr2.FieldName)
	}
}

// TestTagParseError_Error tests the Error() method formatting for TagParseError
func TestTagParseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *TagParseError
		expected string
	}{
		{
			name: "empty tag",
			err: &TagParseError{
				FieldName: "DatabaseURL",
				TagKey:    "config",
				Issue:     "empty config tag",
			},
			expected: "tag parse error in field 'DatabaseURL' (tag: config): empty config tag",
		},
		{
			name: "missing availableAs",
			err: &TagParseError{
				FieldName: "Port",
				TagKey:    "config",
				Issue:     "availableAs not found in config tag",
			},
			expected: "tag parse error in field 'Port' (tag: config): availableAs not found in config tag",
		},
		{
			name: "invalid variable name",
			err: &TagParseError{
				FieldName: "APIKey",
				TagKey:    "config",
				Issue:     "variable name 'api-key!' contains invalid characters",
			},
			expected: "tag parse error in field 'APIKey' (tag: config): variable name 'api-key!' contains invalid characters",
		},
		{
			name: "unknown field name",
			err: &TagParseError{
				FieldName: "<unknown>",
				TagKey:    "config",
				Issue:     "empty availableAs value",
			},
			expected: "tag parse error in field '<unknown>' (tag: config): empty availableAs value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestTagParseError_As tests that errors.As can extract TagParseError from error chains
func TestTagParseError_As(t *testing.T) {
	tagErr := &TagParseError{
		FieldName: "DatabaseURL",
		TagKey:    "config",
		Issue:     "empty config tag",
	}

	// Test direct error
	var extractedErr *TagParseError
	if !errors.As(tagErr, &extractedErr) {
		t.Error("errors.As failed to extract TagParseError")
	}

	if extractedErr.FieldName != "DatabaseURL" {
		t.Errorf("FieldName = %v, want DatabaseURL", extractedErr.FieldName)
	}
	if extractedErr.TagKey != "config" {
		t.Errorf("TagKey = %v, want config", extractedErr.TagKey)
	}
	if extractedErr.Issue != "empty config tag" {
		t.Errorf("Issue = %v, want empty config tag", extractedErr.Issue)
	}

	// Test wrapped error
	wrappedErr := fmt.Errorf("wrapper: %w", tagErr)
	var extractedErr2 *TagParseError
	if !errors.As(wrappedErr, &extractedErr2) {
		t.Error("errors.As failed to extract TagParseError from wrapped error")
	}

	if extractedErr2.FieldName != "DatabaseURL" {
		t.Errorf("FieldName = %v, want DatabaseURL", extractedErr2.FieldName)
	}
}

// TestDependencyGraphError_Error tests the Error() method formatting for DependencyGraphError
func TestDependencyGraphError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *DependencyGraphError
		expected string
	}{
		{
			name: "topological sort failure",
			err: &DependencyGraphError{
				Operation: "topological sort",
				Message:   "unable to complete sort: possible cycle",
			},
			expected: "dependency graph error during topological sort: unable to complete sort: possible cycle",
		},
		{
			name: "build graph failure",
			err: &DependencyGraphError{
				Operation: "build graph",
				Message:   "invalid dependency structure",
			},
			expected: "dependency graph error during build graph: invalid dependency structure",
		},
		{
			name: "generic failure",
			err: &DependencyGraphError{
				Operation: "analyze dependencies",
				Message:   "unexpected error occurred",
			},
			expected: "dependency graph error during analyze dependencies: unexpected error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestDependencyGraphError_As tests that errors.As can extract DependencyGraphError from error chains
func TestDependencyGraphError_As(t *testing.T) {
	graphErr := &DependencyGraphError{
		Operation: "topological sort",
		Message:   "unable to complete sort",
	}

	// Test direct error
	var extractedErr *DependencyGraphError
	if !errors.As(graphErr, &extractedErr) {
		t.Error("errors.As failed to extract DependencyGraphError")
	}

	if extractedErr.Operation != "topological sort" {
		t.Errorf("Operation = %v, want topological sort", extractedErr.Operation)
	}
	if extractedErr.Message != "unable to complete sort" {
		t.Errorf("Message = %v, want unable to complete sort", extractedErr.Message)
	}

	// Test wrapped error
	wrappedErr := fmt.Errorf("wrapper: %w", graphErr)
	var extractedErr2 *DependencyGraphError
	if !errors.As(wrappedErr, &extractedErr2) {
		t.Error("errors.As failed to extract DependencyGraphError from wrapped error")
	}

	if extractedErr2.Operation != "topological sort" {
		t.Errorf("Operation = %v, want topological sort", extractedErr2.Operation)
	}
}

// TestErrorCreationWithVariousFieldCombinations tests creating errors with different field combinations
func TestErrorCreationWithVariousFieldCombinations(t *testing.T) {
	t.Run("LoaderError with all fields", func(t *testing.T) {
		err := &LoaderError{
			LoaderType: "JSONLoader",
			Operation:  "read file",
			Source:     "config.json",
			Err:        fmt.Errorf("file not found"),
		}
		if err.LoaderType != "JSONLoader" {
			t.Errorf("LoaderType not set correctly")
		}
		if err.Operation != "read file" {
			t.Errorf("Operation not set correctly")
		}
		if err.Source != "config.json" {
			t.Errorf("Source not set correctly")
		}
		if err.Err == nil {
			t.Errorf("Err not set correctly")
		}
	})

	t.Run("LoaderError without source", func(t *testing.T) {
		err := &LoaderError{
			LoaderType: "EnvironmentLoader",
			Operation:  "parse",
			Err:        fmt.Errorf("parse error"),
		}
		if err.Source != "" {
			t.Errorf("Source should be empty")
		}
		if err.Error() == "" {
			t.Errorf("Error message should not be empty")
		}
	})

	t.Run("ValidationError with all fields", func(t *testing.T) {
		err := &ValidationError{
			FieldName: "Port",
			Rule:      "min=1",
			Value:     "0",
			Err:       fmt.Errorf("validation failed"),
		}
		if err.FieldName != "Port" {
			t.Errorf("FieldName not set correctly")
		}
		if err.Rule != "min=1" {
			t.Errorf("Rule not set correctly")
		}
		if err.Value != "0" {
			t.Errorf("Value not set correctly")
		}
		if err.Err == nil {
			t.Errorf("Err not set correctly")
		}
	})

	t.Run("ValidationError without value", func(t *testing.T) {
		err := &ValidationError{
			FieldName: "Email",
			Rule:      "required",
			Err:       fmt.Errorf("validation failed"),
		}
		if err.Value != "" {
			t.Errorf("Value should be empty")
		}
		if err.Error() == "" {
			t.Errorf("Error message should not be empty")
		}
	})

	t.Run("TagParseError with all fields", func(t *testing.T) {
		err := &TagParseError{
			FieldName: "DatabaseURL",
			TagKey:    "config",
			Issue:     "empty config tag",
		}
		if err.FieldName != "DatabaseURL" {
			t.Errorf("FieldName not set correctly")
		}
		if err.TagKey != "config" {
			t.Errorf("TagKey not set correctly")
		}
		if err.Issue != "empty config tag" {
			t.Errorf("Issue not set correctly")
		}
	})

	t.Run("DependencyGraphError with all fields", func(t *testing.T) {
		err := &DependencyGraphError{
			Operation: "topological sort",
			Message:   "unable to complete sort",
		}
		if err.Operation != "topological sort" {
			t.Errorf("Operation not set correctly")
		}
		if err.Message != "unable to complete sort" {
			t.Errorf("Message not set correctly")
		}
	})
}

// TestErrorChainTraversal tests that error chains can be traversed correctly
func TestErrorChainTraversal(t *testing.T) {
	t.Run("LoaderError chain", func(t *testing.T) {
		baseErr := fmt.Errorf("base error")
		loaderErr := &LoaderError{
			LoaderType: "JSONLoader",
			Operation:  "unmarshal",
			Err:        baseErr,
		}
		wrappedErr := fmt.Errorf("wrapper: %w", loaderErr)

		// Test errors.As can extract LoaderError
		var extractedLoader *LoaderError
		if !errors.As(wrappedErr, &extractedLoader) {
			t.Error("Failed to extract LoaderError from chain")
		}

		// Test we can still access the base error
		if !errors.Is(wrappedErr, baseErr) {
			t.Error("Failed to find base error in chain")
		}
	})

	t.Run("ValidationError chain", func(t *testing.T) {
		baseErr := fmt.Errorf("base validation error")
		validationErr := &ValidationError{
			FieldName: "Port",
			Rule:      "min=1",
			Err:       baseErr,
		}
		wrappedErr := fmt.Errorf("wrapper: %w", validationErr)

		// Test errors.As can extract ValidationError
		var extractedValidation *ValidationError
		if !errors.As(wrappedErr, &extractedValidation) {
			t.Error("Failed to extract ValidationError from chain")
		}

		// Test we can still access the base error
		if !errors.Is(wrappedErr, baseErr) {
			t.Error("Failed to find base error in chain")
		}
	})

	t.Run("Multiple custom errors in chain", func(t *testing.T) {
		baseErr := fmt.Errorf("base error")
		loaderErr := &LoaderError{
			LoaderType: "JSONLoader",
			Operation:  "read file",
			Err:        baseErr,
		}
		validationErr := &ValidationError{
			FieldName: "Config",
			Rule:      "required",
			Err:       loaderErr,
		}

		// Test we can extract both error types
		var extractedValidation *ValidationError
		if !errors.As(validationErr, &extractedValidation) {
			t.Error("Failed to extract ValidationError")
		}

		var extractedLoader *LoaderError
		if !errors.As(validationErr, &extractedLoader) {
			t.Error("Failed to extract LoaderError from chain")
		}

		// Test we can still access the base error
		if !errors.Is(validationErr, baseErr) {
			t.Error("Failed to find base error in chain")
		}
	})
}
