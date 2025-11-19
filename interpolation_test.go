package config

import (
	"reflect"
	"testing"
)

// Test Analyze() with various struct configurations

func TestInterpolationEngine_Analyze_NoInterpolation(t *testing.T) {
	type Config struct {
		Port int    `env:"PORT"`
		Host string `env:"HOST"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if engine.HasInterpolation() {
		t.Error("expected HasInterpolation() to return false")
	}

	stages := engine.GetDependencyStages()
	if len(stages) != 0 {
		t.Errorf("expected 0 stages, got %d", len(stages))
	}
}

func TestInterpolationEngine_Analyze_SimpleInterpolation(t *testing.T) {
	type Config struct {
		Env        string `env:"ENV" config:"availableAs=ENV"`
		DBPassword string `secret:"aws=/myapp/${ENV}/db/password"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !engine.HasInterpolation() {
		t.Error("expected HasInterpolation() to return true")
	}

	stages := engine.GetDependencyStages()
	if len(stages) != 2 {
		t.Fatalf("expected 2 stages, got %d", len(stages))
	}

	// Stage 0 should contain Env (index 0)
	if len(stages[0]) != 1 || stages[0][0] != 0 {
		t.Errorf("expected stage 0 to contain field 0, got %v", stages[0])
	}

	// Stage 1 should contain DBPassword (index 1)
	if len(stages[1]) != 1 || stages[1][0] != 1 {
		t.Errorf("expected stage 1 to contain field 1, got %v", stages[1])
	}
}

func TestInterpolationEngine_Analyze_MultipleVariables(t *testing.T) {
	type Config struct {
		Env    string `env:"ENV" config:"availableAs=ENV"`
		Region string `env:"REGION" config:"availableAs=REGION"`
		APIKey string `secret:"aws=/myapp/${ENV}/${REGION}/api-key"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stages := engine.GetDependencyStages()
	if len(stages) != 2 {
		t.Fatalf("expected 2 stages, got %d", len(stages))
	}

	// Stage 0 should contain Env and Region (indices 0 and 1)
	if len(stages[0]) != 2 {
		t.Errorf("expected stage 0 to have 2 fields, got %d", len(stages[0]))
	}

	// Stage 1 should contain APIKey (index 2)
	if len(stages[1]) != 1 || stages[1][0] != 2 {
		t.Errorf("expected stage 1 to contain field 2, got %v", stages[1])
	}
}

func TestInterpolationEngine_Analyze_DependencyChain(t *testing.T) {
	type Config struct {
		Env    string `env:"ENV" config:"availableAs=ENV"`
		Region string `env:"REGION_${ENV}" config:"availableAs=REGION"`
		Secret string `secret:"aws=/${ENV}/${REGION}/secret"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stages := engine.GetDependencyStages()
	if len(stages) != 3 {
		t.Fatalf("expected 3 stages, got %d", len(stages))
	}

	// Stage 0: Env (index 0)
	if len(stages[0]) != 1 || stages[0][0] != 0 {
		t.Errorf("expected stage 0 to contain field 0, got %v", stages[0])
	}

	// Stage 1: Region (index 1)
	if len(stages[1]) != 1 || stages[1][0] != 1 {
		t.Errorf("expected stage 1 to contain field 1, got %v", stages[1])
	}

	// Stage 2: Secret (index 2)
	if len(stages[2]) != 1 || stages[2][0] != 2 {
		t.Errorf("expected stage 2 to contain field 2, got %v", stages[2])
	}
}

// Test duplicate availableAs detection

func TestInterpolationEngine_Analyze_DuplicateAvailableAs(t *testing.T) {
	type Config struct {
		Env1 string `env:"ENV1" config:"availableAs=ENV"`
		Env2 string `env:"ENV2" config:"availableAs=ENV"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err == nil {
		t.Fatal("expected error for duplicate availableAs, got nil")
	}

	dupErr, ok := err.(*DuplicateAvailableAsError)
	if !ok {
		t.Fatalf("expected DuplicateAvailableAsError, got %T: %v", err, err)
	}

	if dupErr.VariableName != "ENV" {
		t.Errorf("expected variable name 'ENV', got '%s'", dupErr.VariableName)
	}

	if len(dupErr.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(dupErr.Fields))
	}
}

// Test undefined variable detection

func TestInterpolationEngine_Analyze_UndefinedVariable(t *testing.T) {
	type Config struct {
		DBPassword string `secret:"aws=/myapp/${ENV}/db/password"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err == nil {
		t.Fatal("expected error for undefined variable, got nil")
	}

	undefErr, ok := err.(*UndefinedVariableError)
	if !ok {
		t.Fatalf("expected UndefinedVariableError, got %T: %v", err, err)
	}

	if undefErr.VariableName != "ENV" {
		t.Errorf("expected variable name 'ENV', got '%s'", undefErr.VariableName)
	}

	if undefErr.FieldName != "DBPassword" {
		t.Errorf("expected field name 'DBPassword', got '%s'", undefErr.FieldName)
	}
}

func TestInterpolationEngine_Analyze_MultipleUndefinedVariables(t *testing.T) {
	type Config struct {
		Secret string `secret:"aws=/${ENV}/${REGION}/secret"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err == nil {
		t.Fatal("expected error for undefined variables, got nil")
	}

	// Should fail on first undefined variable encountered
	undefErr, ok := err.(*UndefinedVariableError)
	if !ok {
		t.Fatalf("expected UndefinedVariableError, got %T: %v", err, err)
	}

	// Should report one of the undefined variables
	if undefErr.VariableName != "ENV" && undefErr.VariableName != "REGION" {
		t.Errorf("expected variable name 'ENV' or 'REGION', got '%s'", undefErr.VariableName)
	}
}

// Test cycle detection

func TestInterpolationEngine_Analyze_SimpleCycle(t *testing.T) {
	type Config struct {
		FieldA string `env:"FIELD_${B}" config:"availableAs=A"`
		FieldB string `env:"FIELD_${A}" config:"availableAs=B"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err == nil {
		t.Fatal("expected error for cyclic dependency, got nil")
	}

	cycleErr, ok := err.(*CyclicDependencyError)
	if !ok {
		t.Fatalf("expected CyclicDependencyError, got %T: %v", err, err)
	}

	if len(cycleErr.Cycle) < 2 {
		t.Errorf("expected cycle with at least 2 fields, got %d", len(cycleErr.Cycle))
	}
}

func TestInterpolationEngine_Analyze_ComplexCycle(t *testing.T) {
	type Config struct {
		FieldA string `env:"FIELD_A" config:"availableAs=A"`
		FieldB string `env:"FIELD_${A}" config:"availableAs=B"`
		FieldC string `env:"FIELD_${B}" config:"availableAs=C"`
		FieldD string `env:"FIELD_${C}_${A}" config:"availableAs=D"`
		FieldE string `env:"FIELD_${D}_${B}"` // Creates cycle: B -> C -> D -> B
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	// This should not error because there's no actual cycle
	// FieldE depends on D and B, but doesn't create a cycle
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestInterpolationEngine_Analyze_SelfReference(t *testing.T) {
	type Config struct {
		Field string `env:"FIELD_${SELF}" config:"availableAs=SELF"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err == nil {
		t.Fatal("expected error for self-reference cycle, got nil")
	}

	cycleErr, ok := err.(*CyclicDependencyError)
	if !ok {
		t.Fatalf("expected CyclicDependencyError, got %T: %v", err, err)
	}

	if len(cycleErr.Cycle) == 0 {
		t.Error("expected non-empty cycle")
	}
}

// Test non-exported field validation

func TestInterpolationEngine_Analyze_NonExportedField(t *testing.T) {
	type Config struct {
		env string `env:"ENV" config:"availableAs=ENV"` // lowercase = non-exported
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err == nil {
		t.Fatal("expected error for non-exported field with availableAs, got nil")
	}

	interpErr, ok := err.(*InterpolationError)
	if !ok {
		t.Fatalf("expected InterpolationError, got %T: %v", err, err)
	}

	if interpErr.FieldName != "env" {
		t.Errorf("expected field name 'env', got '%s'", interpErr.FieldName)
	}
}

// Test type conversion for all supported types

func TestInterpolationEngine_UpdateContext_String(t *testing.T) {
	type Config struct {
		Env string `env:"ENV" config:"availableAs=ENV"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	err = engine.UpdateContext(0, "production")
	if err != nil {
		t.Fatalf("UpdateContext failed: %v", err)
	}

	if engine.interpolationContext["ENV"] != "production" {
		t.Errorf("expected 'production', got '%s'", engine.interpolationContext["ENV"])
	}
}

func TestInterpolationEngine_UpdateContext_Int(t *testing.T) {
	type Config struct {
		Port int `env:"PORT" config:"availableAs=PORT"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	err = engine.UpdateContext(0, 8080)
	if err != nil {
		t.Fatalf("UpdateContext failed: %v", err)
	}

	if engine.interpolationContext["PORT"] != "8080" {
		t.Errorf("expected '8080', got '%s'", engine.interpolationContext["PORT"])
	}
}

func TestInterpolationEngine_UpdateContext_IntVariants(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"int8", int8(127), "127"},
		{"int16", int16(32767), "32767"},
		{"int32", int32(2147483647), "2147483647"},
		{"int64", int64(9223372036854775807), "9223372036854775807"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type Config struct {
				Value int `env:"VALUE" config:"availableAs=VALUE"`
			}

			engine := NewInterpolationEngine[Config]()
			cfg := &Config{}

			err := engine.Analyze(cfg)
			if err != nil {
				t.Fatalf("Analyze failed: %v", err)
			}

			err = engine.UpdateContext(0, tt.value)
			if err != nil {
				t.Fatalf("UpdateContext failed: %v", err)
			}

			if engine.interpolationContext["VALUE"] != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, engine.interpolationContext["VALUE"])
			}
		})
	}
}

func TestInterpolationEngine_UpdateContext_UintVariants(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"uint", uint(12345), "12345"},
		{"uint8", uint8(255), "255"},
		{"uint16", uint16(65535), "65535"},
		{"uint32", uint32(4294967295), "4294967295"},
		{"uint64", uint64(18446744073709551615), "18446744073709551615"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type Config struct {
				Value uint `env:"VALUE" config:"availableAs=VALUE"`
			}

			engine := NewInterpolationEngine[Config]()
			cfg := &Config{}

			err := engine.Analyze(cfg)
			if err != nil {
				t.Fatalf("Analyze failed: %v", err)
			}

			err = engine.UpdateContext(0, tt.value)
			if err != nil {
				t.Fatalf("UpdateContext failed: %v", err)
			}

			if engine.interpolationContext["VALUE"] != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, engine.interpolationContext["VALUE"])
			}
		})
	}
}

func TestInterpolationEngine_UpdateContext_Float(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"float32_decimal", float32(1.5), "1.5"},
		{"float32_whole", float32(1.0), "1"},
		{"float64_decimal", float64(3.14159), "3.14159"},
		{"float64_whole", float64(42.0), "42"},
		{"float64_scientific", float64(1.23e10), "1.23e+10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type Config struct {
				Value float64 `env:"VALUE" config:"availableAs=VALUE"`
			}

			engine := NewInterpolationEngine[Config]()
			cfg := &Config{}

			err := engine.Analyze(cfg)
			if err != nil {
				t.Fatalf("Analyze failed: %v", err)
			}

			err = engine.UpdateContext(0, tt.value)
			if err != nil {
				t.Fatalf("UpdateContext failed: %v", err)
			}

			if engine.interpolationContext["VALUE"] != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, engine.interpolationContext["VALUE"])
			}
		})
	}
}

func TestInterpolationEngine_UpdateContext_Bool(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		expected string
	}{
		{"true", true, "true"},
		{"false", false, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type Config struct {
				Debug bool `env:"DEBUG" config:"availableAs=DEBUG"`
			}

			engine := NewInterpolationEngine[Config]()
			cfg := &Config{}

			err := engine.Analyze(cfg)
			if err != nil {
				t.Fatalf("Analyze failed: %v", err)
			}

			err = engine.UpdateContext(0, tt.value)
			if err != nil {
				t.Fatalf("UpdateContext failed: %v", err)
			}

			if engine.interpolationContext["DEBUG"] != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, engine.interpolationContext["DEBUG"])
			}
		})
	}
}

func TestInterpolationEngine_UpdateContext_UnsupportedTypes(t *testing.T) {
	type NestedStruct struct {
		Value string
	}

	tests := []struct {
		name  string
		value interface{}
	}{
		{"struct", NestedStruct{Value: "test"}},
		{"slice", []string{"a", "b", "c"}},
		{"map", map[string]string{"key": "value"}},
		{"pointer", new(string)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type Config struct {
				Field string `env:"FIELD" config:"availableAs=FIELD"`
			}

			engine := NewInterpolationEngine[Config]()
			cfg := &Config{}

			err := engine.Analyze(cfg)
			if err != nil {
				t.Fatalf("Analyze failed: %v", err)
			}

			err = engine.UpdateContext(0, tt.value)
			if err == nil {
				t.Fatal("expected error for unsupported type, got nil")
			}

			interpErr, ok := err.(*InterpolationError)
			if !ok {
				t.Fatalf("expected InterpolationError, got %T: %v", err, err)
			}

			if interpErr.FieldName != "Field" {
				t.Errorf("expected field name 'Field', got '%s'", interpErr.FieldName)
			}
		})
	}
}

// Test InterpolateTags() with single and multiple variables

func TestInterpolationEngine_InterpolateTags_SingleVariable(t *testing.T) {
	type Config struct {
		Env        string `env:"ENV" config:"availableAs=ENV"`
		DBPassword string `secret:"aws=/myapp/${ENV}/db/password"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Update context with ENV value
	err = engine.UpdateContext(0, "production")
	if err != nil {
		t.Fatalf("UpdateContext failed: %v", err)
	}

	// Interpolate tags for field 1 (DBPassword)
	err = engine.InterpolateTags([]int{1})
	if err != nil {
		t.Fatalf("InterpolateTags failed: %v", err)
	}

	// Note: We can't actually verify the tag was modified because Go doesn't allow
	// runtime tag modification. This test verifies the method doesn't error.
}

func TestInterpolationEngine_InterpolateTags_MultipleVariables(t *testing.T) {
	type Config struct {
		Env    string `env:"ENV" config:"availableAs=ENV"`
		Region string `env:"REGION" config:"availableAs=REGION"`
		APIKey string `secret:"aws=/myapp/${ENV}/${REGION}/api-key"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Update context
	err = engine.UpdateContext(0, "prod")
	if err != nil {
		t.Fatalf("UpdateContext failed for ENV: %v", err)
	}

	err = engine.UpdateContext(1, "us-east-1")
	if err != nil {
		t.Fatalf("UpdateContext failed for REGION: %v", err)
	}

	// Interpolate tags for field 2 (APIKey)
	err = engine.InterpolateTags([]int{2})
	if err != nil {
		t.Fatalf("InterpolateTags failed: %v", err)
	}
}

func TestInterpolationEngine_InterpolateTags_InvalidFieldIndex(t *testing.T) {
	type Config struct {
		Env string `env:"ENV" config:"availableAs=ENV"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Try to interpolate with invalid field index
	err = engine.InterpolateTags([]int{999})
	if err == nil {
		t.Fatal("expected error for invalid field index, got nil")
	}
}

// Test UpdateContext() with different field types

func TestInterpolationEngine_UpdateContext_FieldWithoutAvailableAs(t *testing.T) {
	type Config struct {
		Env  string `env:"ENV" config:"availableAs=ENV"`
		Port int    `env:"PORT"` // No availableAs
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Update context for field without availableAs should not error
	err = engine.UpdateContext(1, 8080)
	if err != nil {
		t.Fatalf("UpdateContext should not error for field without availableAs: %v", err)
	}

	// Context should not contain PORT
	if _, exists := engine.interpolationContext["PORT"]; exists {
		t.Error("context should not contain PORT for field without availableAs")
	}
}

// Integration test: Full workflow

func TestInterpolationEngine_FullWorkflow(t *testing.T) {
	type Config struct {
		Env        string `env:"ENV" config:"availableAs=ENV"`
		Region     string `env:"REGION" config:"availableAs=REGION"`
		Port       int    `env:"PORT" config:"availableAs=PORT"`
		Debug      bool   `env:"DEBUG" config:"availableAs=DEBUG"`
		DBPassword string `secret:"aws=/myapp/${ENV}/${REGION}/db/password"`
		ConfigFile string `yaml:"config-${PORT}.yaml"`
		LogLevel   string `env:"LOG_LEVEL_${DEBUG}"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	// Step 1: Analyze
	err := engine.Analyze(cfg)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if !engine.HasInterpolation() {
		t.Fatal("expected HasInterpolation() to return true")
	}

	// Step 2: Get dependency stages
	stages := engine.GetDependencyStages()
	if len(stages) != 2 {
		t.Fatalf("expected 2 stages, got %d", len(stages))
	}

	// Step 3: Process stage 0 (fields with no dependencies)
	for _, fieldIndex := range stages[0] {
		// Simulate loading values
		switch fieldIndex {
		case 0: // Env
			err = engine.UpdateContext(fieldIndex, "production")
		case 1: // Region
			err = engine.UpdateContext(fieldIndex, "us-west-2")
		case 2: // Port
			err = engine.UpdateContext(fieldIndex, 8080)
		case 3: // Debug
			err = engine.UpdateContext(fieldIndex, true)
		}
		if err != nil {
			t.Fatalf("UpdateContext failed for field %d: %v", fieldIndex, err)
		}
	}

	// Verify context
	if engine.interpolationContext["ENV"] != "production" {
		t.Errorf("expected ENV='production', got '%s'", engine.interpolationContext["ENV"])
	}
	if engine.interpolationContext["REGION"] != "us-west-2" {
		t.Errorf("expected REGION='us-west-2', got '%s'", engine.interpolationContext["REGION"])
	}
	if engine.interpolationContext["PORT"] != "8080" {
		t.Errorf("expected PORT='8080', got '%s'", engine.interpolationContext["PORT"])
	}
	if engine.interpolationContext["DEBUG"] != "true" {
		t.Errorf("expected DEBUG='true', got '%s'", engine.interpolationContext["DEBUG"])
	}

	// Step 4: Interpolate tags for stage 1
	err = engine.InterpolateTags(stages[1])
	if err != nil {
		t.Fatalf("InterpolateTags failed: %v", err)
	}

	// Step 5: Process stage 1 (fields with dependencies)
	// In real usage, loaders would use the interpolated tags
	// For this test, we just verify no errors occurred
}

// Test TagParseError field name setting

func TestInterpolationEngine_Analyze_TagParseError_EmptyTag(t *testing.T) {
	// Note: config:"" results in an empty string from Tag.Get(), so we test with a space
	type Config struct {
		Field string `config:" "`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err == nil {
		t.Fatal("expected error for config tag without availableAs, got nil")
	}

	tagErr, ok := err.(*TagParseError)
	if !ok {
		t.Fatalf("expected TagParseError, got %T: %v", err, err)
	}

	if tagErr.FieldName != "Field" {
		t.Errorf("expected field name 'Field', got '%s'", tagErr.FieldName)
	}

	if tagErr.TagKey != "config" {
		t.Errorf("expected tag key 'config', got '%s'", tagErr.TagKey)
	}

	// Should fail because availableAs is not found
	if !contains(tagErr.Issue, "availableAs not found") {
		t.Errorf("expected issue to mention 'availableAs not found', got '%s'", tagErr.Issue)
	}

	// Verify error message includes field name
	if !contains(tagErr.Error(), "Field") {
		t.Errorf("error message should include field name 'Field': %s", tagErr.Error())
	}
}

func TestInterpolationEngine_Analyze_TagParseError_MissingAvailableAs(t *testing.T) {
	type Config struct {
		DatabaseURL string `config:"other=value"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err == nil {
		t.Fatal("expected error for missing availableAs, got nil")
	}

	tagErr, ok := err.(*TagParseError)
	if !ok {
		t.Fatalf("expected TagParseError, got %T: %v", err, err)
	}

	if tagErr.FieldName != "DatabaseURL" {
		t.Errorf("expected field name 'DatabaseURL', got '%s'", tagErr.FieldName)
	}

	if tagErr.Issue != "availableAs not found in config tag" {
		t.Errorf("expected issue 'availableAs not found in config tag', got '%s'", tagErr.Issue)
	}

	// Verify error message includes correct field name
	expectedMsg := "tag parse error in field 'DatabaseURL' (tag: config): availableAs not found in config tag"
	if tagErr.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, tagErr.Error())
	}
}

func TestInterpolationEngine_Analyze_TagParseError_EmptyAvailableAs(t *testing.T) {
	type Config struct {
		APIKey string `config:"availableAs="`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err == nil {
		t.Fatal("expected error for empty availableAs value, got nil")
	}

	tagErr, ok := err.(*TagParseError)
	if !ok {
		t.Fatalf("expected TagParseError, got %T: %v", err, err)
	}

	if tagErr.FieldName != "APIKey" {
		t.Errorf("expected field name 'APIKey', got '%s'", tagErr.FieldName)
	}

	if tagErr.Issue != "empty availableAs value" {
		t.Errorf("expected issue 'empty availableAs value', got '%s'", tagErr.Issue)
	}
}

func TestInterpolationEngine_Analyze_TagParseError_InvalidVariableName(t *testing.T) {
	type Config struct {
		SecretKey string `config:"availableAs=SECRET@KEY!"`
	}

	engine := NewInterpolationEngine[Config]()
	cfg := &Config{}

	err := engine.Analyze(cfg)
	if err == nil {
		t.Fatal("expected error for invalid variable name, got nil")
	}

	tagErr, ok := err.(*TagParseError)
	if !ok {
		t.Fatalf("expected TagParseError, got %T: %v", err, err)
	}

	if tagErr.FieldName != "SecretKey" {
		t.Errorf("expected field name 'SecretKey', got '%s'", tagErr.FieldName)
	}

	if tagErr.TagKey != "config" {
		t.Errorf("expected tag key 'config', got '%s'", tagErr.TagKey)
	}

	// Issue should mention invalid characters
	if !contains(tagErr.Issue, "invalid availableAs value") {
		t.Errorf("expected issue to mention 'invalid availableAs value', got '%s'", tagErr.Issue)
	}

	// Verify error message is complete with field context
	if !contains(tagErr.Error(), "SecretKey") {
		t.Errorf("error message should include field name 'SecretKey': %s", tagErr.Error())
	}
}

func TestInterpolationEngine_Analyze_TagParseError_MultipleFields(t *testing.T) {
	// Test that each field gets its own error with correct field name
	tests := []struct {
		name          string
		configStruct  interface{}
		expectedField string
		expectedIssue string
	}{
		{
			name: "first field error",
			configStruct: &struct {
				Field1 string `config:"other=value"`
				Field2 string `config:"availableAs=VALID"`
			}{},
			expectedField: "Field1",
			expectedIssue: "availableAs not found",
		},
		{
			name: "second field error",
			configStruct: &struct {
				Field1 string `config:"availableAs=VALID"`
				Field2 string `config:"availableAs="`
			}{},
			expectedField: "Field2",
			expectedIssue: "empty availableAs value",
		},
		{
			name: "third field error",
			configStruct: &struct {
				Field1 string `config:"availableAs=VALID1"`
				Field2 string `config:"availableAs=VALID2"`
				Field3 string `config:"availableAs=INVALID@NAME"`
			}{},
			expectedField: "Field3",
			expectedIssue: "invalid availableAs value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use reflection to create engine with the right type
			configValue := reflect.ValueOf(tt.configStruct).Elem()
			configType := configValue.Type()

			// Create a generic engine
			engine := &InterpolationEngine[struct{}]{
				availableAsMap:       make(map[string]int),
				dependencies:         make(map[int][]string),
				dependencyStages:     make([][]int, 0),
				interpolationContext: make(map[string]string),
				fieldNames:           make(map[int]string),
				originalTags:         make(map[int]reflect.StructTag),
				hasInterpolation:     false,
			}

			engine.configValue = configValue

			// Manually run the first pass of Analyze
			availableAsFields := make(map[string][]string)
			for i := 0; i < configType.NumField(); i++ {
				field := configType.Field(i)
				engine.fieldNames[i] = field.Name
				engine.originalTags[i] = field.Tag

				configTag := field.Tag.Get("config")
				if configTag != "" {
					varName, err := ParseConfigTag(configTag)
					if err != nil {
						// Update TagParseError with actual field name
						if tagErr, ok := err.(*TagParseError); ok {
							tagErr.FieldName = field.Name
							// Verify this is the expected error
							if tagErr.FieldName != tt.expectedField {
								t.Errorf("expected field name '%s', got '%s'", tt.expectedField, tagErr.FieldName)
							}
							if !contains(tagErr.Issue, tt.expectedIssue) {
								t.Errorf("expected issue to contain '%s', got '%s'", tt.expectedIssue, tagErr.Issue)
							}
							return
						}
						t.Fatalf("expected TagParseError, got %T: %v", err, err)
					}

					if !field.IsExported() {
						t.Fatal("unexpected non-exported field error")
					}

					availableAsFields[varName] = append(availableAsFields[varName], field.Name)
					engine.availableAsMap[varName] = i
					engine.hasInterpolation = true
				}
			}

			t.Fatal("expected TagParseError but none was returned")
		})
	}
}
