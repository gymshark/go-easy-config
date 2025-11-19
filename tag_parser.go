package config

import (
	"fmt"
	"regexp"
	"strings"
)

// Variable reference pattern: ${VAR_NAME} where VAR_NAME contains alphanumeric, underscore, or hyphen
var variableReferenceRegex = regexp.MustCompile(`\$\{([A-Za-z0-9_-]+)\}`)

// ParseConfigTag extracts the availableAs value from a config struct tag.
// Returns the variable name and nil error if found, or empty string and TagParseError if not found or malformed.
//
// Example:
//
//	ParseConfigTag(`config:"availableAs=ENV"`) returns ("ENV", nil)
//	ParseConfigTag(`config:"other=value"`) returns ("", TagParseError)
//
// Note: The returned TagParseError will have FieldName set to "<unknown>".
// The caller (e.g., InterpolationEngine) should update this with the actual field name.
func ParseConfigTag(tag string) (string, error) {
	if tag == "" {
		return "", &TagParseError{
			FieldName: "<unknown>",
			TagKey:    "config",
			Issue:     "empty config tag",
		}
	}

	// Look for availableAs=VALUE pattern
	prefix := "availableAs="
	idx := strings.Index(tag, prefix)
	if idx == -1 {
		return "", &TagParseError{
			FieldName: "<unknown>",
			TagKey:    "config",
			Issue:     "availableAs not found in config tag",
		}
	}

	// Extract value after availableAs=
	start := idx + len(prefix)
	value := tag[start:]

	// Handle comma-separated attributes (take only the first part)
	if commaIdx := strings.Index(value, ","); commaIdx != -1 {
		value = value[:commaIdx]
	}

	// Handle space-separated attributes
	if spaceIdx := strings.Index(value, " "); spaceIdx != -1 {
		value = value[:spaceIdx]
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", &TagParseError{
			FieldName: "<unknown>",
			TagKey:    "config",
			Issue:     "empty availableAs value",
		}
	}

	// Validate the variable name
	if err := ValidateVariableName(value); err != nil {
		return "", &TagParseError{
			FieldName: "<unknown>",
			TagKey:    "config",
			Issue:     fmt.Sprintf("invalid availableAs value: %v", err),
		}
	}

	return value, nil
}

// FindVariableReferences extracts all ${VAR} references from a string.
// Returns a slice of variable names (without the ${} syntax).
// Duplicate variable names are included multiple times if they appear multiple times.
//
// Example:
//
//	FindVariableReferences("path/${ENV}/file") returns []string{"ENV"}
//	FindVariableReferences("${VAR1}/${VAR2}") returns []string{"VAR1", "VAR2"}
//	FindVariableReferences("${VAR}${VAR}") returns []string{"VAR", "VAR"}
func FindVariableReferences(s string) []string {
	matches := variableReferenceRegex.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil
	}

	vars := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			vars = append(vars, match[1]) // Extract variable name from capture group
		}
	}
	return vars
}

// InterpolateString replaces all ${VAR} references in a string with values from the context map.
// Returns the interpolated string and nil error if all variables are found.
// Returns an error if any variable is undefined in the context.
//
// Example:
//
//	context := map[string]string{"ENV": "prod", "REGION": "us-east-1"}
//	InterpolateString("/app/${ENV}/${REGION}/config", context) returns ("/app/prod/us-east-1/config", nil)
//	InterpolateString("${MISSING}", context) returns ("", error)
func InterpolateString(s string, context map[string]string) (string, error) {
	var missingVars []string

	result := variableReferenceRegex.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name from ${VAR}
		varName := match[2 : len(match)-1]

		if value, ok := context[varName]; ok {
			return value
		}

		// Track missing variables for error reporting
		missingVars = append(missingVars, varName)
		return match // Keep original if not found
	})

	if len(missingVars) > 0 {
		return "", fmt.Errorf("undefined variables: %v", missingVars)
	}

	return result, nil
}

// ValidateVariableName checks if a variable name follows the allowed pattern.
// Variable names must contain only alphanumeric characters, underscores, and hyphens.
// Empty names are not allowed.
//
// Example:
//
//	ValidateVariableName("ENV") returns nil
//	ValidateVariableName("MY_VAR-123") returns nil
//	ValidateVariableName("") returns error
//	ValidateVariableName("VAR@NAME") returns error
func ValidateVariableName(name string) error {
	if name == "" {
		return fmt.Errorf("variable name cannot be empty")
	}

	// Check if name matches allowed pattern
	validNameRegex := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	if !validNameRegex.MatchString(name) {
		return fmt.Errorf("variable name '%s' contains invalid characters (only alphanumeric, underscore, and hyphen allowed)", name)
	}

	return nil
}
