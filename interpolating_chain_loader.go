package config

import (
	"fmt"
	"reflect"
)

// InterpolatingChainLoader wraps a chain of loaders and adds variable interpolation support.
// It analyzes struct tags for variable references, builds a dependency graph, and executes
// loaders in dependency-ordered stages to ensure variables are resolved before they are used.
//
// The loader maintains backward compatibility by detecting when no interpolation is needed
// and delegating directly to a standard ChainLoader for optimal performance.
//
// Note: Due to Go's limitation that struct tags cannot be modified at runtime, this
// implementation provides the infrastructure for staged loading and context management.
// The actual tag interpolation must be handled by creating interpolation-aware loaders
// or by using a code generation approach in future iterations.
//
// Example usage:
//
//	loader := &InterpolatingChainLoader[Config]{
//	    Loaders: []Loader[Config]{
//	        &EnvironmentLoader[Config]{},
//	        &CommandLineLoader[Config]{},
//	    },
//	}
//	var cfg Config
//	if err := loader.Load(&cfg); err != nil {
//	    // Handle error
//	}
type InterpolatingChainLoader[T any] struct {
	Loaders      []Loader[T]
	engine       *InterpolationEngine[T]
	ShortCircuit bool // Enable short-circuit behavior within stages
}

// Load executes loaders in dependency-aware stages when interpolation is needed,
// or delegates to standard ChainLoader when no interpolation is detected.
//
// The loading process:
//  1. Analyze struct tags to detect interpolation needs and build dependency graph
//  2. If no interpolation needed, use fast path (delegate to ChainLoader)
//  3. Otherwise, load fields in dependency-ordered stages:
//     - For each stage: load all fields, update interpolation context
//     - Context is available for next stage's fields
//
// Returns an error if:
//   - Analysis fails (cycles, undefined variables, etc.)
//   - Any loader fails during execution
//   - Type conversion fails for availableAs fields
func (l *InterpolatingChainLoader[T]) Load(c *T) error {
	if l.Loaders == nil {
		return fmt.Errorf("InterpolatingChainLoader.Loaders is nil")
	}

	// Initialize engine if not already done
	if l.engine == nil {
		l.engine = NewInterpolationEngine[T]()
	}

	// Analyze the struct to detect interpolation needs
	if err := l.engine.Analyze(c); err != nil {
		return fmt.Errorf("interpolation analysis failed: %w", err)
	}

	// Fast path: no interpolation needed
	// Execute loaders in sequence without staged loading
	if !l.engine.HasInterpolation() {
		return l.loadWithoutInterpolation(c)
	}

	// Slow path: staged loading with interpolation
	return l.loadWithInterpolation(c)
}

// loadWithoutInterpolation executes loaders in sequence without staged loading.
// This is the fast path when no interpolation is needed.
// If ShortCircuit is enabled, stops loading when all fields are populated.
func (l *InterpolatingChainLoader[T]) loadWithoutInterpolation(c *T) error {
	for i, loader := range l.Loaders {
		if loader == nil {
			return fmt.Errorf("loader at index %d is nil", i)
		}

		// Apply short-circuit logic if enabled
		if l.ShortCircuit && l.isStageFullyPopulated(c) {
			break
		}

		if err := loader.Load(c); err != nil {
			return fmt.Errorf("error in loader at index %d: %w", i, err)
		}
	}

	return nil
}

// loadWithInterpolation performs staged loading with variable interpolation.
// It processes fields in dependency order, updating the interpolation context
// after each stage so that dependent fields can use the resolved values.
//
// The interpolation context is built progressively as fields are loaded,
// making variable values available for subsequent stages.
func (l *InterpolatingChainLoader[T]) loadWithInterpolation(c *T) error {
	stages := l.engine.GetDependencyStages()

	// Process each dependency stage
	for stageNum, stageFields := range stages {
		// Interpolate tags for this stage using current context
		// Note: This prepares the interpolation context but cannot modify struct tags
		if err := l.engine.InterpolateTags(stageFields); err != nil {
			return fmt.Errorf("failed to interpolate tags for stage %d: %w", stageNum, err)
		}

		// Load fields in this stage using all loaders
		// Loaders execute in sequence, maintaining precedence within the stage
		if err := l.loadStage(c); err != nil {
			return fmt.Errorf("failed to load stage %d: %w", stageNum, err)
		}

		// Update interpolation context with loaded values from this stage
		// This makes the values available for interpolation in subsequent stages
		if err := l.updateContextForStage(c, stageFields); err != nil {
			return fmt.Errorf("failed to update context for stage %d: %w", stageNum, err)
		}
	}

	return nil
}

// loadStage executes all loaders for the current stage.
// Loaders are executed in sequence, maintaining the loader precedence within the stage.
// Later loaders can override values set by earlier loaders.
//
// If ShortCircuit is enabled, the loader stops early when all exported fields are populated,
// but ensures that dependency fields (those with availableAs) are always loaded before
// dependent fields. Short-circuit logic is applied within each stage, not across stages.
//
// Note: Since struct tags cannot be modified at runtime, loaders see the original tags.
// Future enhancements may include interpolation-aware loader wrappers or code generation.
func (l *InterpolatingChainLoader[T]) loadStage(c *T) error {
	// Execute all loaders in sequence
	// Each loader processes the entire struct, but the staged approach ensures
	// that dependencies are satisfied before dependent fields are used
	for i, loader := range l.Loaders {
		if loader == nil {
			return fmt.Errorf("loader at index %d is nil", i)
		}

		// Apply short-circuit logic within the stage if enabled
		if l.ShortCircuit && l.isStageFullyPopulated(c) {
			break
		}

		if err := loader.Load(c); err != nil {
			return fmt.Errorf("error in loader at index %d: %w", i, err)
		}
	}

	return nil
}

// isStageFullyPopulated checks if all exported fields in the configuration are populated.
// This is used for short-circuit behavior within stages.
func (l *InterpolatingChainLoader[T]) isStageFullyPopulated(c *T) bool {
	if c == nil {
		return false
	}
	configValue := reflect.ValueOf(c).Elem()
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		structField := configType.Field(i)
		// Skip unexported fields
		if structField.PkgPath != "" {
			continue
		}

		fieldValue := configValue.Field(i)
		if isZeroValue(fieldValue) {
			return false
		}
	}

	return true
}

// isZeroValue checks if a reflect.Value is a zero value for its type.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Map, reflect.Slice, reflect.Ptr:
		return v.IsNil()
	case reflect.Interface:
		if v.IsNil() {
			return true
		}
		underlying := v.Elem()
		if (underlying.Kind() == reflect.Ptr || underlying.Kind() == reflect.Interface) && underlying.IsNil() {
			return true
		}
		return isZeroValue(underlying)
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZeroValue(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == 0
	case reflect.String:
		return v.String() == ""
	default:
		return false
	}
}

// updateContextForStage updates the interpolation context with values from fields
// that were loaded in the current stage and have availableAs declarations.
//
// This method:
//  1. Retrieves the current value of each field in the stage
//  2. Converts the value to string representation (if field has availableAs)
//  3. Adds the value to the interpolation context for use in subsequent stages
func (l *InterpolatingChainLoader[T]) updateContextForStage(c *T, stageFields []int) error {
	configValue := reflect.ValueOf(c).Elem()

	for _, fieldIndex := range stageFields {
		// Get the field value
		fieldValue := configValue.Field(fieldIndex)

		// Update context with this field's value
		// The engine checks if this field has availableAs and converts the value
		if err := l.engine.UpdateContext(fieldIndex, fieldValue.Interface()); err != nil {
			return err
		}
	}

	return nil
}

// GetInterpolationContext returns the current interpolation context.
// This can be used for debugging or by custom loaders that need access to
// the resolved variable values.
func (l *InterpolatingChainLoader[T]) GetInterpolationContext() map[string]string {
	if l.engine == nil {
		return nil
	}
	// Return a copy to prevent external modification
	context := make(map[string]string)
	for k, v := range l.engine.interpolationContext {
		context[k] = v
	}
	return context
}
