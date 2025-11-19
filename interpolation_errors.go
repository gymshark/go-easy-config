package config

import (
	"fmt"
	"strings"
)

// InterpolationError represents errors during variable interpolation.
// It includes context about which field encountered the error and provides
// a descriptive message to aid in debugging configuration issues.
//
// Fields:
//   - FieldName: Name of the field where interpolation failed (e.g., "DatabaseURL", "SecretPath")
//   - Message: Descriptive error message explaining what went wrong
//
// Operations that return InterpolationError:
//   - InterpolationEngine.Analyze() - When interpolation analysis fails
//   - InterpolatingChainLoader.Load() - When variable interpolation fails during loading
//
// Example - Inspecting interpolation errors:
//
//	handler := config.NewConfigHandler[AppConfig]()
//	var cfg AppConfig
//	if err := handler.Load(&cfg); err != nil {
//	    var interpErr *InterpolationError
//	    if errors.As(err, &interpErr) {
//	        fmt.Printf("Interpolation failed in field '%s'\n", interpErr.FieldName)
//	        fmt.Printf("Error: %s\n", interpErr.Message)
//	    }
//	}
type InterpolationError struct {
	FieldName string
	Message   string
}

// Error implements the error interface for InterpolationError.
// Returns a formatted error message including the field name and specific error details.
func (e *InterpolationError) Error() string {
	return fmt.Sprintf("interpolation error in field '%s': %s", e.FieldName, e.Message)
}

// CyclicDependencyError represents circular dependency detection in field dependencies.
// It includes the complete cycle path to help identify which fields are involved
// in the circular reference.
//
// Fields:
//   - Cycle: Ordered list of field names forming the circular dependency
//
// Operations that return CyclicDependencyError:
//   - DependencyGraph.TopologicalSort() - When a cycle is detected during topological sort
//   - DependencyGraph.DetectCycle() - When explicitly checking for cycles
//
// Example - Inspecting cyclic dependency errors:
//
//	handler := config.NewConfigHandler[AppConfig]()
//	var cfg AppConfig
//	if err := handler.Load(&cfg); err != nil {
//	    var cycleErr *CyclicDependencyError
//	    if errors.As(err, &cycleErr) {
//	        fmt.Printf("Circular dependency detected!\n")
//	        fmt.Printf("Cycle: %s\n", strings.Join(cycleErr.Cycle, " -> "))
//	        fmt.Printf("To fix: Break the dependency chain by restructuring your configuration\n")
//	    }
//	}
//
// Example scenario that causes this error:
//
//	type Config struct {
//	    // FieldA depends on FieldB
//	    FieldA string `env:"A_${B}" config:"availableAs=A"`
//	    // FieldB depends on FieldA - creates a cycle!
//	    FieldB string `env:"B_${A}" config:"availableAs=B"`
//	}
//	// Error: cyclic dependency detected: FieldA -> FieldB -> FieldA
type CyclicDependencyError struct {
	Cycle []string // Field names in the cycle
}

// Error implements the error interface for CyclicDependencyError.
// Returns a formatted error message showing the complete dependency cycle.
func (e *CyclicDependencyError) Error() string {
	return fmt.Sprintf("cyclic dependency detected: %s", strings.Join(e.Cycle, " -> "))
}

// UndefinedVariableError represents reference to a non-existent variable.
// It includes both the field making the reference and the variable name
// that was not found in the availableAs declarations.
//
// Fields:
//   - FieldName: Name of the field that references the undefined variable
//   - VariableName: Name of the variable that was not found
//
// Operations that return UndefinedVariableError:
//   - BuildDependencyGraph() - When a referenced variable has no corresponding availableAs declaration
//
// Example - Inspecting undefined variable errors:
//
//	handler := config.NewConfigHandler[AppConfig]()
//	var cfg AppConfig
//	if err := handler.Load(&cfg); err != nil {
//	    var undefErr *UndefinedVariableError
//	    if errors.As(err, &undefErr) {
//	        fmt.Printf("Undefined variable '${%s}' in field '%s'\n",
//	            undefErr.VariableName, undefErr.FieldName)
//	        fmt.Printf("To fix: Add config:\"availableAs=%s\" to the field providing this value\n",
//	            undefErr.VariableName)
//	    }
//	}
//
// Example scenario that causes this error:
//
//	type Config struct {
//	    // References ${ENV} but no field declares availableAs=ENV
//	    DatabaseURL string `secret:"aws=/myapp/${ENV}/db/password"`
//	}
//	// Error: undefined variable '${ENV}' referenced in field 'DatabaseURL'
//	//
//	// Fix by adding:
//	// Environment string `env:"ENV" config:"availableAs=ENV"`
type UndefinedVariableError struct {
	FieldName    string
	VariableName string
}

// Error implements the error interface for UndefinedVariableError.
// Returns a formatted error message indicating which variable is undefined
// and where it was referenced.
func (e *UndefinedVariableError) Error() string {
	return fmt.Sprintf("undefined variable '${%s}' referenced in field '%s'", e.VariableName, e.FieldName)
}

// DuplicateAvailableAsError represents duplicate variable declarations.
// It includes the variable name and all fields that declared it,
// helping identify which declarations need to be renamed.
//
// Fields:
//   - VariableName: Name of the variable that was declared multiple times
//   - Fields: List of field names that declared the same variable
//
// Operations that return DuplicateAvailableAsError:
//   - InterpolationEngine.Analyze() - When multiple fields declare the same availableAs name
//
// Example - Inspecting duplicate variable errors:
//
//	handler := config.NewConfigHandler[AppConfig]()
//	var cfg AppConfig
//	if err := handler.Load(&cfg); err != nil {
//	    var dupErr *DuplicateAvailableAsError
//	    if errors.As(err, &dupErr) {
//	        fmt.Printf("Duplicate variable '%s' declared in fields: %s\n",
//	            dupErr.VariableName, strings.Join(dupErr.Fields, ", "))
//	        fmt.Printf("To fix: Use unique names for each availableAs declaration\n")
//	    }
//	}
//
// Example scenario that causes this error:
//
//	type Config struct {
//	    // Both fields declare availableAs=ENV - not allowed!
//	    Environment string `env:"ENV" config:"availableAs=ENV"`
//	    EnvName     string `env:"ENVIRONMENT" config:"availableAs=ENV"`
//	}
//	// Error: duplicate availableAs='ENV' declared in fields: Environment, EnvName
//	//
//	// Fix by using unique names:
//	// Environment string `env:"ENV" config:"availableAs=ENV"`
//	// EnvName     string `env:"ENVIRONMENT" config:"availableAs=ENV_NAME"`
type DuplicateAvailableAsError struct {
	VariableName string
	Fields       []string // Field names with duplicate declarations
}

// Error implements the error interface for DuplicateAvailableAsError.
// Returns a formatted error message listing all fields with the duplicate declaration.
func (e *DuplicateAvailableAsError) Error() string {
	return fmt.Sprintf("duplicate availableAs='%s' declared in fields: %s", e.VariableName, strings.Join(e.Fields, ", "))
}
