package config

import (
	"strings"
	"testing"
)

func TestInterpolationError_Error(t *testing.T) {
	tests := []struct {
		name      string
		err       *InterpolationError
		wantError string
	}{
		{
			name: "basic interpolation error",
			err: &InterpolationError{
				FieldName: "DBPassword",
				Message:   "unsupported type for interpolation: struct",
			},
			wantError: "interpolation error in field 'DBPassword': unsupported type for interpolation: struct",
		},
		{
			name: "type conversion error",
			err: &InterpolationError{
				FieldName: "Config",
				Message:   "cannot convert complex type to string",
			},
			wantError: "interpolation error in field 'Config': cannot convert complex type to string",
		},
		{
			name: "malformed syntax error",
			err: &InterpolationError{
				FieldName: "APIKey",
				Message:   "malformed variable syntax: ${INCOMPLETE",
			},
			wantError: "interpolation error in field 'APIKey': malformed variable syntax: ${INCOMPLETE",
		},
		{
			name: "non-exported field error",
			err: &InterpolationError{
				FieldName: "privateField",
				Message:   "field with availableAs must be exported (starts with uppercase)",
			},
			wantError: "interpolation error in field 'privateField': field with availableAs must be exported (starts with uppercase)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantError {
				t.Errorf("InterpolationError.Error() = %q, want %q", got, tt.wantError)
			}
		})
	}
}

func TestCyclicDependencyError_Error(t *testing.T) {
	tests := []struct {
		name      string
		err       *CyclicDependencyError
		wantError string
	}{
		{
			name: "simple two-field cycle",
			err: &CyclicDependencyError{
				Cycle: []string{"FieldA", "FieldB", "FieldA"},
			},
			wantError: "cyclic dependency detected: FieldA -> FieldB -> FieldA",
		},
		{
			name: "three-field cycle",
			err: &CyclicDependencyError{
				Cycle: []string{"FieldA", "FieldB", "FieldC", "FieldA"},
			},
			wantError: "cyclic dependency detected: FieldA -> FieldB -> FieldC -> FieldA",
		},
		{
			name: "self-referencing cycle",
			err: &CyclicDependencyError{
				Cycle: []string{"FieldA", "FieldA"},
			},
			wantError: "cyclic dependency detected: FieldA -> FieldA",
		},
		{
			name: "complex multi-field cycle",
			err: &CyclicDependencyError{
				Cycle: []string{"Environment", "Region", "ConfigPath", "Environment"},
			},
			wantError: "cyclic dependency detected: Environment -> Region -> ConfigPath -> Environment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantError {
				t.Errorf("CyclicDependencyError.Error() = %q, want %q", got, tt.wantError)
			}
		})
	}
}

func TestUndefinedVariableError_Error(t *testing.T) {
	tests := []struct {
		name      string
		err       *UndefinedVariableError
		wantError string
	}{
		{
			name: "undefined environment variable",
			err: &UndefinedVariableError{
				FieldName:    "DBPassword",
				VariableName: "ENV",
			},
			wantError: "undefined variable '${ENV}' referenced in field 'DBPassword'",
		},
		{
			name: "undefined region variable",
			err: &UndefinedVariableError{
				FieldName:    "APIKey",
				VariableName: "REGION",
			},
			wantError: "undefined variable '${REGION}' referenced in field 'APIKey'",
		},
		{
			name: "typo in variable name",
			err: &UndefinedVariableError{
				FieldName:    "ConfigPath",
				VariableName: "ENVIORNMENT",
			},
			wantError: "undefined variable '${ENVIORNMENT}' referenced in field 'ConfigPath'",
		},
		{
			name: "variable with hyphen",
			err: &UndefinedVariableError{
				FieldName:    "SecretPath",
				VariableName: "APP-NAME",
			},
			wantError: "undefined variable '${APP-NAME}' referenced in field 'SecretPath'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantError {
				t.Errorf("UndefinedVariableError.Error() = %q, want %q", got, tt.wantError)
			}
		})
	}
}

func TestDuplicateAvailableAsError_Error(t *testing.T) {
	tests := []struct {
		name      string
		err       *DuplicateAvailableAsError
		wantError string
	}{
		{
			name: "duplicate in two fields",
			err: &DuplicateAvailableAsError{
				VariableName: "ENV",
				Fields:       []string{"Environment", "Env"},
			},
			wantError: "duplicate availableAs='ENV' declared in fields: Environment, Env",
		},
		{
			name: "duplicate in three fields",
			err: &DuplicateAvailableAsError{
				VariableName: "REGION",
				Fields:       []string{"Region", "RegionName", "AWSRegion"},
			},
			wantError: "duplicate availableAs='REGION' declared in fields: Region, RegionName, AWSRegion",
		},
		{
			name: "duplicate with hyphenated name",
			err: &DuplicateAvailableAsError{
				VariableName: "APP-NAME",
				Fields:       []string{"ApplicationName", "AppName"},
			},
			wantError: "duplicate availableAs='APP-NAME' declared in fields: ApplicationName, AppName",
		},
		{
			name: "duplicate with underscore name",
			err: &DuplicateAvailableAsError{
				VariableName: "DB_HOST",
				Fields:       []string{"DatabaseHost", "DBHost"},
			},
			wantError: "duplicate availableAs='DB_HOST' declared in fields: DatabaseHost, DBHost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantError {
				t.Errorf("DuplicateAvailableAsError.Error() = %q, want %q", got, tt.wantError)
			}
		})
	}
}

// TestErrorTypes verifies that all error types implement the error interface
func TestErrorTypes(t *testing.T) {
	var _ error = &InterpolationError{}
	var _ error = &CyclicDependencyError{}
	var _ error = &UndefinedVariableError{}
	var _ error = &DuplicateAvailableAsError{}
}

// TestErrorContextInformation verifies that errors contain sufficient context
func TestErrorContextInformation(t *testing.T) {
	t.Run("InterpolationError contains field name", func(t *testing.T) {
		err := &InterpolationError{
			FieldName: "TestField",
			Message:   "test message",
		}
		errMsg := err.Error()
		if !strings.Contains(errMsg, "TestField") {
			t.Errorf("InterpolationError message should contain field name 'TestField', got: %s", errMsg)
		}
	})

	t.Run("CyclicDependencyError contains all cycle fields", func(t *testing.T) {
		err := &CyclicDependencyError{
			Cycle: []string{"Field1", "Field2", "Field3"},
		}
		errMsg := err.Error()
		for _, field := range err.Cycle {
			if !strings.Contains(errMsg, field) {
				t.Errorf("CyclicDependencyError message should contain field '%s', got: %s", field, errMsg)
			}
		}
	})

	t.Run("UndefinedVariableError contains both field and variable", func(t *testing.T) {
		err := &UndefinedVariableError{
			FieldName:    "TestField",
			VariableName: "TEST_VAR",
		}
		errMsg := err.Error()
		if !strings.Contains(errMsg, "TestField") {
			t.Errorf("UndefinedVariableError message should contain field name 'TestField', got: %s", errMsg)
		}
		if !strings.Contains(errMsg, "TEST_VAR") {
			t.Errorf("UndefinedVariableError message should contain variable name 'TEST_VAR', got: %s", errMsg)
		}
	})

	t.Run("DuplicateAvailableAsError contains variable and all fields", func(t *testing.T) {
		err := &DuplicateAvailableAsError{
			VariableName: "TEST_VAR",
			Fields:       []string{"Field1", "Field2"},
		}
		errMsg := err.Error()
		if !strings.Contains(errMsg, "TEST_VAR") {
			t.Errorf("DuplicateAvailableAsError message should contain variable name 'TEST_VAR', got: %s", errMsg)
		}
		for _, field := range err.Fields {
			if !strings.Contains(errMsg, field) {
				t.Errorf("DuplicateAvailableAsError message should contain field '%s', got: %s", field, errMsg)
			}
		}
	})
}
