package config

import (
	"errors"
	"testing"
)

func TestParseConfigTag(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		wantVar     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "simple availableAs",
			tag:     `availableAs=ENV`,
			wantVar: "ENV",
			wantErr: false,
		},
		{
			name:    "availableAs with other attributes",
			tag:     `availableAs=REGION,other=value`,
			wantVar: "REGION",
			wantErr: false,
		},
		{
			name:    "availableAs with spaces",
			tag:     `availableAs=PORT other=value`,
			wantVar: "PORT",
			wantErr: false,
		},
		{
			name:    "availableAs with underscore",
			tag:     `availableAs=MY_VAR`,
			wantVar: "MY_VAR",
			wantErr: false,
		},
		{
			name:    "availableAs with hyphen",
			tag:     `availableAs=MY-VAR`,
			wantVar: "MY-VAR",
			wantErr: false,
		},
		{
			name:    "availableAs with numbers",
			tag:     `availableAs=VAR123`,
			wantVar: "VAR123",
			wantErr: false,
		},
		{
			name:        "empty tag",
			tag:         "",
			wantErr:     true,
			errContains: "empty config tag",
		},
		{
			name:        "no availableAs",
			tag:         `other=value`,
			wantErr:     true,
			errContains: "availableAs not found",
		},
		{
			name:        "empty availableAs value",
			tag:         `availableAs=`,
			wantErr:     true,
			errContains: "empty availableAs value",
		},
		{
			name:        "availableAs with invalid characters",
			tag:         `availableAs=VAR@NAME`,
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:    "availableAs with spaces in name",
			tag:     `availableAs=MY VAR`,
			wantVar: "MY",
			wantErr: false, // Space acts as delimiter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVar, err := ParseConfigTag(tt.tag)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseConfigTag() expected error but got none")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ParseConfigTag() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseConfigTag() unexpected error = %v", err)
				return
			}

			if gotVar != tt.wantVar {
				t.Errorf("ParseConfigTag() = %v, want %v", gotVar, tt.wantVar)
			}
		})
	}
}

func TestFindVariableReferences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantVars []string
	}{
		{
			name:     "no variables",
			input:    "plain text",
			wantVars: nil,
		},
		{
			name:     "single variable",
			input:    "${ENV}",
			wantVars: []string{"ENV"},
		},
		{
			name:     "variable in path",
			input:    "/app/${ENV}/config",
			wantVars: []string{"ENV"},
		},
		{
			name:     "multiple variables",
			input:    "${VAR1}/${VAR2}",
			wantVars: []string{"VAR1", "VAR2"},
		},
		{
			name:     "adjacent variables",
			input:    "${VAR1}${VAR2}",
			wantVars: []string{"VAR1", "VAR2"},
		},
		{
			name:     "duplicate variables",
			input:    "${VAR}/${VAR}",
			wantVars: []string{"VAR", "VAR"},
		},
		{
			name:     "variable with underscore",
			input:    "${MY_VAR}",
			wantVars: []string{"MY_VAR"},
		},
		{
			name:     "variable with hyphen",
			input:    "${MY-VAR}",
			wantVars: []string{"MY-VAR"},
		},
		{
			name:     "variable with numbers",
			input:    "${VAR123}",
			wantVars: []string{"VAR123"},
		},
		{
			name:     "complex path with multiple variables",
			input:    "/myapp/${ENV}/${REGION}/db/${DB_NAME}",
			wantVars: []string{"ENV", "REGION", "DB_NAME"},
		},
		{
			name:     "empty braces",
			input:    "${}",
			wantVars: nil, // Empty name doesn't match pattern
		},
		{
			name:     "unclosed brace",
			input:    "${VAR",
			wantVars: nil, // Doesn't match pattern
		},
		{
			name:     "no opening brace",
			input:    "VAR}",
			wantVars: nil,
		},
		{
			name:     "dollar sign without braces",
			input:    "$VAR",
			wantVars: nil,
		},
		{
			name:     "nested braces",
			input:    "${{VAR}}",
			wantVars: nil, // Doesn't match pattern
		},
		{
			name:     "variable with invalid characters",
			input:    "${VAR@NAME}",
			wantVars: nil, // @ not allowed in pattern
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVars := FindVariableReferences(tt.input)

			if !equalStringSlices(gotVars, tt.wantVars) {
				t.Errorf("FindVariableReferences() = %v, want %v", gotVars, tt.wantVars)
			}
		})
	}
}

func TestInterpolateString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		context     map[string]string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "no variables",
			input:   "plain text",
			context: map[string]string{},
			want:    "plain text",
			wantErr: false,
		},
		{
			name:    "single variable",
			input:   "${ENV}",
			context: map[string]string{"ENV": "prod"},
			want:    "prod",
			wantErr: false,
		},
		{
			name:    "variable in path",
			input:   "/app/${ENV}/config",
			context: map[string]string{"ENV": "prod"},
			want:    "/app/prod/config",
			wantErr: false,
		},
		{
			name:    "multiple variables",
			input:   "${VAR1}/${VAR2}",
			context: map[string]string{"VAR1": "value1", "VAR2": "value2"},
			want:    "value1/value2",
			wantErr: false,
		},
		{
			name:    "adjacent variables",
			input:   "${VAR1}${VAR2}",
			context: map[string]string{"VAR1": "hello", "VAR2": "world"},
			want:    "helloworld",
			wantErr: false,
		},
		{
			name:    "duplicate variables",
			input:   "${VAR}/${VAR}",
			context: map[string]string{"VAR": "test"},
			want:    "test/test",
			wantErr: false,
		},
		{
			name:    "complex path",
			input:   "/myapp/${ENV}/${REGION}/db/${DB_NAME}",
			context: map[string]string{"ENV": "prod", "REGION": "us-east-1", "DB_NAME": "main"},
			want:    "/myapp/prod/us-east-1/db/main",
			wantErr: false,
		},
		{
			name:    "value with path separator",
			input:   "/app/${PATH}/config",
			context: map[string]string{"PATH": "sub/dir"},
			want:    "/app/sub/dir/config",
			wantErr: false,
		},
		{
			name:        "undefined variable",
			input:       "${MISSING}",
			context:     map[string]string{"ENV": "prod"},
			wantErr:     true,
			errContains: "undefined variables",
		},
		{
			name:        "multiple undefined variables",
			input:       "${VAR1}/${VAR2}",
			context:     map[string]string{},
			wantErr:     true,
			errContains: "undefined variables",
		},
		{
			name:        "partial context",
			input:       "${VAR1}/${VAR2}",
			context:     map[string]string{"VAR1": "value1"},
			wantErr:     true,
			errContains: "undefined variables",
		},
		{
			name:    "empty context with no variables",
			input:   "no variables here",
			context: map[string]string{},
			want:    "no variables here",
			wantErr: false,
		},
		{
			name:    "variable with underscore",
			input:   "${MY_VAR}",
			context: map[string]string{"MY_VAR": "value"},
			want:    "value",
			wantErr: false,
		},
		{
			name:    "variable with hyphen",
			input:   "${MY-VAR}",
			context: map[string]string{"MY-VAR": "value"},
			want:    "value",
			wantErr: false,
		},
		{
			name:    "variable with numbers",
			input:   "${VAR123}",
			context: map[string]string{"VAR123": "value"},
			want:    "value",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InterpolateString(tt.input, tt.context)

			if tt.wantErr {
				if err == nil {
					t.Errorf("InterpolateString() expected error but got none")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("InterpolateString() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("InterpolateString() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("InterpolateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateVariableName(t *testing.T) {
	tests := []struct {
		name        string
		varName     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid simple name",
			varName: "ENV",
			wantErr: false,
		},
		{
			name:    "valid with underscore",
			varName: "MY_VAR",
			wantErr: false,
		},
		{
			name:    "valid with hyphen",
			varName: "MY-VAR",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			varName: "VAR123",
			wantErr: false,
		},
		{
			name:    "valid mixed case",
			varName: "MyVar",
			wantErr: false,
		},
		{
			name:    "valid complex",
			varName: "MY_VAR-123",
			wantErr: false,
		},
		{
			name:        "empty name",
			varName:     "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "with space",
			varName:     "MY VAR",
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "with at sign",
			varName:     "VAR@NAME",
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "with dollar sign",
			varName:     "$VAR",
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "with dot",
			varName:     "VAR.NAME",
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "with slash",
			varName:     "VAR/NAME",
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "with brackets",
			varName:     "VAR[0]",
			wantErr:     true,
			errContains: "invalid characters",
		},
		{
			name:        "with braces",
			varName:     "{VAR}",
			wantErr:     true,
			errContains: "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVariableName(tt.varName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateVariableName() expected error but got none")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateVariableName() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateVariableName() unexpected error = %v", err)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function to compare string slices
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestParseConfigTag_ReturnsTagParseError(t *testing.T) {
	tests := []struct {
		name          string
		tag           string
		wantFieldName string
		wantTagKey    string
		wantIssue     string
	}{
		{
			name:          "empty tag",
			tag:           "",
			wantFieldName: "<unknown>",
			wantTagKey:    "config",
			wantIssue:     "empty config tag",
		},
		{
			name:          "no availableAs",
			tag:           "other=value",
			wantFieldName: "<unknown>",
			wantTagKey:    "config",
			wantIssue:     "availableAs not found in config tag",
		},
		{
			name:          "empty availableAs value",
			tag:           "availableAs=",
			wantFieldName: "<unknown>",
			wantTagKey:    "config",
			wantIssue:     "empty availableAs value",
		},
		{
			name:          "invalid variable name with special characters",
			tag:           "availableAs=VAR@NAME",
			wantFieldName: "<unknown>",
			wantTagKey:    "config",
			wantIssue:     "invalid availableAs value",
		},
		{
			name:          "invalid variable name with dollar sign",
			tag:           "availableAs=$VAR",
			wantFieldName: "<unknown>",
			wantTagKey:    "config",
			wantIssue:     "invalid availableAs value",
		},
		{
			name:          "invalid variable name with dot",
			tag:           "availableAs=VAR.NAME",
			wantFieldName: "<unknown>",
			wantTagKey:    "config",
			wantIssue:     "invalid availableAs value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseConfigTag(tt.tag)

			if err == nil {
				t.Fatalf("ParseConfigTag() expected error but got none")
			}

			// Check if error is TagParseError
			var tagErr *TagParseError
			if !errors.As(err, &tagErr) {
				t.Fatalf("ParseConfigTag() error type = %T, want *TagParseError", err)
			}

			// Verify FieldName is set to <unknown>
			if tagErr.FieldName != tt.wantFieldName {
				t.Errorf("TagParseError.FieldName = %q, want %q", tagErr.FieldName, tt.wantFieldName)
			}

			// Verify TagKey is set to "config"
			if tagErr.TagKey != tt.wantTagKey {
				t.Errorf("TagParseError.TagKey = %q, want %q", tagErr.TagKey, tt.wantTagKey)
			}

			// Verify Issue contains expected text
			if !contains(tagErr.Issue, tt.wantIssue) {
				t.Errorf("TagParseError.Issue = %q, want to contain %q", tagErr.Issue, tt.wantIssue)
			}

			// Verify error message format
			errMsg := tagErr.Error()
			if !contains(errMsg, "tag parse error") {
				t.Errorf("TagParseError.Error() = %q, want to contain 'tag parse error'", errMsg)
			}
			if !contains(errMsg, tt.wantTagKey) {
				t.Errorf("TagParseError.Error() = %q, want to contain tag key %q", errMsg, tt.wantTagKey)
			}
			if !contains(errMsg, tt.wantIssue) {
				t.Errorf("TagParseError.Error() = %q, want to contain issue %q", errMsg, tt.wantIssue)
			}
		})
	}
}
