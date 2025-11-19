package config

import (
	"fmt"

	"github.com/gymshark/go-easy-config/loader"
)

// LoaderError is re-exported from the loader package for convenience.
// See loader.LoaderError for full documentation.
type LoaderError = loader.LoaderError

// ValidationError represents validation failures for configuration fields.
// It captures which field failed validation, which rule was violated,
// and optionally the invalid value that caused the failure.
//
// Fields:
//   - FieldName: Name of the field that failed validation (e.g., "Port", "DatabaseURL")
//   - Rule: Validation rule that failed (e.g., "required", "min=1", "url")
//   - Value: Optional string representation of the invalid value
//   - Err: Underlying validator error for accessing detailed validation information
//
// Operations that return ValidationError:
//   - Handler.Validate() - When struct validation fails
//   - Handler.LoadAndValidate() - When validation fails after successful loading
//
// Example - Creating a ValidationError:
//
//	err := &ValidationError{
//	    FieldName: "Port",
//	    Rule:      "min=1",
//	    Value:     "0",
//	    Err:       validatorErr,
//	}
//
// Example - Inspecting validation errors:
//
//	var cfg AppConfig
//	if err := handler.LoadAndValidate(&cfg); err != nil {
//	    var validationErr *ValidationError
//	    if errors.As(err, &validationErr) {
//	        fmt.Printf("Validation failed for field '%s'\n", validationErr.FieldName)
//	        fmt.Printf("Rule '%s' was violated\n", validationErr.Rule)
//	        if validationErr.Value != "" {
//	            fmt.Printf("Invalid value: %s\n", validationErr.Value)
//	        }
//	        // Access underlying validator error for more details
//	        if validationErr.Err != nil {
//	            fmt.Printf("Details: %v\n", validationErr.Err)
//	        }
//	    }
//	}
//
// Note: For multiple validation errors, FieldName and Rule may be set to "<multiple>"
// with the underlying validator error containing all failures. Use errors.Unwrap() or
// access the Err field directly to get the complete validator.ValidationErrors.
type ValidationError struct {
	FieldName string // Name of the field that failed validation
	Rule      string // Validation rule that failed (e.g., "required", "min=1")
	Value     string // Optional string representation of the invalid value
	Err       error  // Underlying validator error
}

// Error returns a formatted error message with validation context.
// If Value is provided, it's included in the message.
func (e *ValidationError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("validation failed for field '%s': rule '%s' failed (value: %s)",
			e.FieldName, e.Rule, e.Value)
	}
	return fmt.Sprintf("validation failed for field '%s': rule '%s' failed",
		e.FieldName, e.Rule)
}

// Unwrap returns the underlying validator error, enabling error chain traversal.
func (e *ValidationError) Unwrap() error {
	return e.Err
}

// TagParseError represents errors that occur when parsing struct tags.
// It captures which field has the problematic tag, which tag key was being parsed,
// and a description of the specific issue encountered.
//
// Fields:
//   - FieldName: Name of the field with the problematic tag (e.g., "DatabaseURL", "Port")
//   - TagKey: Tag key being parsed (e.g., "config", "env", "secret")
//   - Issue: Description of the specific issue (e.g., "empty config tag", "invalid variable name")
//
// Operations that return TagParseError:
//   - ParseConfigTag() - When config tag syntax is invalid
//   - InterpolationEngine.Analyze() - When tag parsing fails during interpolation analysis
//
// Example - Creating a TagParseError:
//
//	err := &TagParseError{
//	    FieldName: "DatabaseURL",
//	    TagKey:    "config",
//	    Issue:     "empty config tag",
//	}
//
// Example - Inspecting tag parse errors:
//
//	handler := config.NewConfigHandler[AppConfig]()
//	var cfg AppConfig
//	if err := handler.Load(&cfg); err != nil {
//	    var tagErr *TagParseError
//	    if errors.As(err, &tagErr) {
//	        fmt.Printf("Tag parse error in field '%s'\n", tagErr.FieldName)
//	        fmt.Printf("Tag key: %s\n", tagErr.TagKey)
//	        fmt.Printf("Issue: %s\n", tagErr.Issue)
//	        // Fix the struct tag based on the issue description
//	    }
//	}
//
// Common Issues:
//   - "empty config tag" - The config tag is present but has no value
//   - "availableAs not found in config tag" - Missing required availableAs parameter
//   - "empty availableAs value" - availableAs parameter is present but empty
//   - "invalid availableAs value" - Variable name contains invalid characters
//
// Note: The tag parser may initially set FieldName to "<unknown>" if it doesn't
// have access to the field name. The caller (e.g., InterpolationEngine) should
// update this field with the actual field name before returning the error.
type TagParseError struct {
	FieldName string // Name of the field with the problematic tag
	TagKey    string // Tag key being parsed (e.g., "config", "env")
	Issue     string // Description of the issue
}

// Error returns a formatted error message with tag parsing context.
func (e *TagParseError) Error() string {
	return fmt.Sprintf("tag parse error in field '%s' (tag: %s): %s",
		e.FieldName, e.TagKey, e.Issue)
}

// DependencyGraphError represents errors that occur during dependency graph operations
// beyond cycles and undefined variables (which have their own specific error types).
// This is used for general dependency graph failures such as topological sort issues.
//
// Fields:
//   - Operation: Operation being performed (e.g., "topological sort", "build graph")
//   - Message: Description of the specific issue encountered
//
// Operations that return DependencyGraphError:
//   - DependencyGraph.TopologicalSort() - When sort fails for reasons other than cycles
//   - BuildDependencyGraph() - When graph construction fails
//
// Example - Creating a DependencyGraphError:
//
//	err := &DependencyGraphError{
//	    Operation: "topological sort",
//	    Message:   "unable to complete sort: possible cycle",
//	}
//
// Example - Inspecting dependency graph errors:
//
//	handler := config.NewConfigHandler[AppConfig]()
//	var cfg AppConfig
//	if err := handler.Load(&cfg); err != nil {
//	    var graphErr *DependencyGraphError
//	    if errors.As(err, &graphErr) {
//	        fmt.Printf("Dependency graph error during %s\n", graphErr.Operation)
//	        fmt.Printf("Details: %s\n", graphErr.Message)
//	        // This indicates a structural issue with field dependencies
//	    }
//	}
//
// Note: For specific dependency issues, see also:
//   - CyclicDependencyError - For circular dependency detection
//   - UndefinedVariableError - For references to non-existent variables
//   - DuplicateAvailableAsError - For duplicate variable declarations
type DependencyGraphError struct {
	Operation string // Operation being performed (e.g., "topological sort", "build graph")
	Message   string // Description of the issue
}

// Error returns a formatted error message with dependency graph context.
func (e *DependencyGraphError) Error() string {
	return fmt.Sprintf("dependency graph error during %s: %s",
		e.Operation, e.Message)
}
