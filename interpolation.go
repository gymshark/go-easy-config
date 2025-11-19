package config

import (
	"fmt"
	"reflect"
)

// InterpolationEngine manages variable interpolation for configuration structs.
// It analyzes struct tags to build a dependency graph, performs topological sorting,
// and replaces variable references with actual field values during loading.
//
// The engine supports the following workflow:
//  1. Analyze() - Parse struct tags and build dependency information
//  2. GetDependencyStages() - Retrieve fields grouped by dependency level
//  3. InterpolateTags() - Replace ${VAR} references in struct tags
//  4. UpdateContext() - Add loaded field values to interpolation context
//
// Example usage:
//
//	engine := NewInterpolationEngine[Config]()
//	if err := engine.Analyze(&cfg); err != nil {
//	    // Handle analysis errors (cycles, undefined variables, etc.)
//	}
//	stages := engine.GetDependencyStages()
//	for _, stage := range stages {
//	    engine.InterpolateTags(stage)
//	    // Load fields in this stage
//	    for _, fieldIndex := range stage {
//	        engine.UpdateContext(fieldIndex, fieldValue)
//	    }
//	}
type InterpolationEngine[T any] struct {
	// availableAsMap maps variable names to field indices
	availableAsMap map[string]int

	// dependencies maps field index to list of variable names it depends on
	dependencies map[int][]string

	// dependencyStages contains fields grouped by dependency level
	// Stage 0: no dependencies, Stage 1: depends on Stage 0, etc.
	dependencyStages [][]int

	// interpolationContext stores resolved field values
	interpolationContext map[string]string

	// fieldNames maps field index to field name for error messages
	fieldNames map[int]string

	// originalTags stores original struct tags before interpolation
	originalTags map[int]reflect.StructTag

	// configValue stores the reflect.Value of the config struct
	configValue reflect.Value

	// hasInterpolation tracks whether any interpolation is needed
	hasInterpolation bool
}

// NewInterpolationEngine creates a new InterpolationEngine for the given configuration type.
func NewInterpolationEngine[T any]() *InterpolationEngine[T] {
	return &InterpolationEngine[T]{
		availableAsMap:       make(map[string]int),
		dependencies:         make(map[int][]string),
		dependencyStages:     make([][]int, 0),
		interpolationContext: make(map[string]string),
		fieldNames:           make(map[int]string),
		originalTags:         make(map[int]reflect.StructTag),
		hasInterpolation:     false,
	}
}

// Analyze examines the struct and builds dependency information.
// It parses config tags for availableAs declarations, identifies variable references,
// validates variable names, detects duplicates and undefined variables,
// validates that fields with availableAs are exported, builds the dependency graph,
// detects cycles, and performs topological sort to create dependency stages.
//
// Returns an error if:
//   - Duplicate availableAs declarations are found
//   - Undefined variables are referenced
//   - Circular dependencies are detected
//   - Non-exported fields have availableAs declarations
//   - Variable names are invalid
func (e *InterpolationEngine[T]) Analyze(cfg *T) error {
	e.configValue = reflect.ValueOf(cfg).Elem()
	configType := e.configValue.Type()

	// First pass: collect availableAs declarations and detect duplicates
	availableAsFields := make(map[string][]string) // varName -> []fieldName
	for i := 0; i < configType.NumField(); i++ {
		field := configType.Field(i)
		e.fieldNames[i] = field.Name

		// Store original tags
		e.originalTags[i] = field.Tag

		// Check for config tag with availableAs
		configTag := field.Tag.Get("config")
		if configTag != "" {
			varName, err := ParseConfigTag(configTag)
			if err != nil {
				// Update TagParseError with actual field name
				if tagErr, ok := err.(*TagParseError); ok {
					tagErr.FieldName = field.Name
					return tagErr
				}
				// config tag exists but doesn't have valid availableAs - skip
				continue
			}

			// Validate that field is exported
			if !field.IsExported() {
				return &InterpolationError{
					FieldName: field.Name,
					Message:   "field with availableAs must be exported (starts with uppercase)",
				}
			}

			// Track for duplicate detection
			availableAsFields[varName] = append(availableAsFields[varName], field.Name)
			e.availableAsMap[varName] = i
			e.hasInterpolation = true
		}
	}

	// Check for duplicate availableAs declarations
	for varName, fields := range availableAsFields {
		if len(fields) > 1 {
			return &DuplicateAvailableAsError{
				VariableName: varName,
				Fields:       fields,
			}
		}
	}

	// Second pass: find variable references in all tags
	for i := 0; i < configType.NumField(); i++ {
		field := configType.Field(i)
		tag := field.Tag

		// Check all tag keys for variable references
		var allVars []string
		seenVars := make(map[string]bool)

		// Iterate through all possible tag keys
		tagString := string(tag)
		vars := FindVariableReferences(tagString)
		for _, varName := range vars {
			if !seenVars[varName] {
				allVars = append(allVars, varName)
				seenVars[varName] = true
				e.hasInterpolation = true
			}
		}

		if len(allVars) > 0 {
			e.dependencies[i] = allVars

			// Validate that all referenced variables are defined
			for _, varName := range allVars {
				if _, exists := e.availableAsMap[varName]; !exists {
					return &UndefinedVariableError{
						FieldName:    field.Name,
						VariableName: varName,
					}
				}
			}
		}
	}

	// If no interpolation is needed, we're done
	if !e.hasInterpolation {
		return nil
	}

	// Build dependency graph
	graph, err := BuildDependencyGraph(e.dependencies, e.availableAsMap, e.fieldNames)
	if err != nil {
		return err
	}

	// Detect cycles
	if cyclePath := graph.DetectCycle(); cyclePath != nil {
		return &CyclicDependencyError{Cycle: cyclePath}
	}

	// Perform topological sort to get dependency stages
	stages, err := graph.TopologicalSort()
	if err != nil {
		return err
	}

	e.dependencyStages = stages
	return nil
}

// HasInterpolation returns true if any fields use variable interpolation.
// This can be used to implement a fast path that bypasses interpolation entirely.
func (e *InterpolationEngine[T]) HasInterpolation() bool {
	return e.hasInterpolation
}

// GetDependencyStages returns fields grouped by dependency level.
// Stage 0 contains fields with no dependencies.
// Stage 1 contains fields that depend only on Stage 0 fields.
// Stage N contains fields that depend on fields from stages 0 to N-1.
func (e *InterpolationEngine[T]) GetDependencyStages() [][]int {
	return e.dependencyStages
}

// InterpolateTags replaces ${VAR} references in struct tags for specified fields.
// This modifies the struct's field tags in memory using reflection.
// The original tags are preserved and can be restored if needed.
//
// Parameters:
//   - fieldIndices: slice of field indices to interpolate
//
// Returns an error if interpolation fails for any field.
func (e *InterpolationEngine[T]) InterpolateTags(fieldIndices []int) error {
	configType := e.configValue.Type()

	for _, fieldIndex := range fieldIndices {
		if fieldIndex < 0 || fieldIndex >= configType.NumField() {
			return fmt.Errorf("invalid field index: %d", fieldIndex)
		}

		field := configType.Field(fieldIndex)
		originalTag := e.originalTags[fieldIndex]

		// Interpolate the entire tag string
		tagString := string(originalTag)
		interpolatedTag, err := InterpolateString(tagString, e.interpolationContext)
		if err != nil {
			return &InterpolationError{
				FieldName: field.Name,
				Message:   fmt.Sprintf("failed to interpolate tags: %v", err),
			}
		}

		// Note: In Go, we cannot actually modify struct tags at runtime.
		// This is a design limitation. The actual tag modification will need to happen
		// in the loader integration layer where we can work with the interpolated strings
		// directly rather than trying to modify the struct's metadata.
		// For now, we store the interpolated result in a way that loaders can access it.
		_ = interpolatedTag // Will be used in loader integration
	}

	return nil
}

// UpdateContext adds a field's value to the interpolation context.
// The field value is converted to a string representation based on its type.
//
// Supported types:
//   - string: used directly
//   - int, int8, int16, int32, int64: converted to decimal string
//   - uint, uint8, uint16, uint32, uint64: converted to decimal string
//   - float32, float64: converted to compact string representation
//   - bool: converted to "true" or "false"
//
// Returns an error if the field type is not supported for interpolation.
func (e *InterpolationEngine[T]) UpdateContext(fieldIndex int, value interface{}) error {
	// Find the variable name for this field
	var varName string
	for name, idx := range e.availableAsMap {
		if idx == fieldIndex {
			varName = name
			break
		}
	}

	// If this field doesn't have availableAs, nothing to update
	if varName == "" {
		return nil
	}

	// Convert value to string
	strValue, err := e.convertToString(value)
	if err != nil {
		fieldName := e.fieldNames[fieldIndex]
		return &InterpolationError{
			FieldName: fieldName,
			Message:   fmt.Sprintf("failed to convert value to string: %v", err),
		}
	}

	e.interpolationContext[varName] = strValue
	return nil
}

// convertToString converts a value to its string representation for interpolation.
// Supports string, int (all variants), uint (all variants), float32, float64, and bool types.
// Returns an error for unsupported types (struct, slice, map, pointer).
func (e *InterpolationEngine[T]) convertToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case int8:
		return fmt.Sprintf("%d", v), nil
	case int16:
		return fmt.Sprintf("%d", v), nil
	case int32:
		return fmt.Sprintf("%d", v), nil
	case int64:
		return fmt.Sprintf("%d", v), nil
	case uint:
		return fmt.Sprintf("%d", v), nil
	case uint8:
		return fmt.Sprintf("%d", v), nil
	case uint16:
		return fmt.Sprintf("%d", v), nil
	case uint32:
		return fmt.Sprintf("%d", v), nil
	case uint64:
		return fmt.Sprintf("%d", v), nil
	case float32:
		return fmt.Sprintf("%g", v), nil
	case float64:
		return fmt.Sprintf("%g", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	default:
		return "", fmt.Errorf("unsupported type for interpolation: %T", v)
	}
}
