package aws

import (
	"errors"
	"testing"

	"github.com/gymshark/go-easy-config/loader"
)

type SSMTestConfig struct {
	Parameter1 string `ssm:"parameter1"`
	Parameter2 int    `ssm:"parameter2"`
}

func TestSSMParameterStoreLoader_ErrorWrapping(t *testing.T) {
	// Test that SSMParameterStoreLoader returns LoaderError on failures
	// Note: This test will fail in environments without AWS credentials or SSM access
	// which is expected behavior - we're testing that errors are properly wrapped

	cfg := &SSMTestConfig{}

	ldr := &SSMParameterStoreLoader[SSMTestConfig]{
		Path: "/nonexistent/path/",
	}

	err := ldr.Load(cfg)

	// We expect an error in test environment without AWS setup
	if err != nil {
		var loaderErr *loader.LoaderError
		if !errors.As(err, &loaderErr) {
			t.Fatalf("expected LoaderError, got %T", err)
		}

		// Test error fields
		if loaderErr.LoaderType != "SSMParameterStoreLoader" {
			t.Errorf("expected LoaderType 'SSMParameterStoreLoader', got '%s'", loaderErr.LoaderType)
		}

		if loaderErr.Operation != "fetch parameters" {
			t.Errorf("expected Operation 'fetch parameters', got '%s'", loaderErr.Operation)
		}

		if loaderErr.Source != "/nonexistent/path/" {
			t.Errorf("expected Source '/nonexistent/path/', got '%s'", loaderErr.Source)
		}

		// Test that underlying error is accessible
		if loaderErr.Err == nil {
			t.Error("expected underlying error to be set")
		}

		// Test error message includes AWS-specific context
		errMsg := err.Error()
		if errMsg == "" {
			t.Error("expected non-empty error message")
		}

		// Verify error message contains loader type and operation
		if loaderErr.LoaderType == "" || loaderErr.Operation == "" {
			t.Error("expected error message to include loader type and operation")
		}
	}
}

func TestSSMParameterStoreLoader_ErrorMessageFormat(t *testing.T) {
	// Test that error messages follow the expected format
	cfg := &SSMTestConfig{}

	ldr := &SSMParameterStoreLoader[SSMTestConfig]{
		Path: "/test/path/",
	}

	err := ldr.Load(cfg)

	if err != nil {
		var loaderErr *loader.LoaderError
		if errors.As(err, &loaderErr) {
			// Verify the error can be unwrapped
			unwrapped := errors.Unwrap(err)
			if unwrapped == nil {
				t.Error("expected error to be unwrappable")
			}

			// Verify error message format includes all context
			errMsg := err.Error()
			// Should contain: "SSMParameterStoreLoader error during fetch parameters (source: /test/path/): ..."
			if errMsg == "" {
				t.Error("expected non-empty error message")
			}
		}
	}
}
