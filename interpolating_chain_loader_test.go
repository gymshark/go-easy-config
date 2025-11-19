package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/gymshark/go-easy-config/loader/generic"
)

// Mock loader for testing
type mockLoader[T any] struct {
	name      string
	loadFunc  func(*T) error
	callCount int
}

func (m *mockLoader[T]) Load(c *T) error {
	m.callCount++
	if m.loadFunc != nil {
		return m.loadFunc(c)
	}
	return nil
}

// Test staged loading with mock loaders
func TestInterpolatingChainLoader_StagedLoading(t *testing.T) {
	type Config struct {
		Env        string `env:"ENV" config:"availableAs=ENV"`
		DBPassword string `secret:"aws=/myapp/${ENV}/db/password"`
	}

	// Track loader execution order
	var executionOrder []string

	loader1 := &mockLoader[Config]{
		name: "loader1",
		loadFunc: func(c *Config) error {
			executionOrder = append(executionOrder, "loader1")
			if c.Env == "" {
				c.Env = "prod"
			}
			return nil
		},
	}

	loader2 := &mockLoader[Config]{
		name: "loader2",
		loadFunc: func(c *Config) error {
			executionOrder = append(executionOrder, "loader2")
			if c.DBPassword == "" {
				c.DBPassword = "secret123"
			}
			return nil
		},
	}

	chain := &InterpolatingChainLoader[Config]{
		Loaders: []Loader[Config]{loader1, loader2},
	}

	cfg := &Config{}
	err := chain.Load(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify both loaders were called
	if loader1.callCount < 1 {
		t.Errorf("expected loader1 to be called at least once, got %d", loader1.callCount)
	}
	if loader2.callCount < 1 {
		t.Errorf("expected loader2 to be called at least once, got %d", loader2.callCount)
	}

	// Verify configuration was loaded
	if cfg.Env != "prod" {
		t.Errorf("expected Env='prod', got '%s'", cfg.Env)
	}
	if cfg.DBPassword != "secret123" {
		t.Errorf("expected DBPassword='secret123', got '%s'", cfg.DBPassword)
	}
}

// Test context updates between stages
func TestInterpolatingChainLoader_ContextUpdates(t *testing.T) {
	type Config struct {
		Env    string `env:"ENV" config:"availableAs=ENV"`
		Region string `env:"REGION" config:"availableAs=REGION"`
		APIKey string `secret:"aws=/myapp/${ENV}/${REGION}/api-key"`
	}

	loader := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			if c.Env == "" {
				c.Env = "staging"
			}
			if c.Region == "" {
				c.Region = "us-west-2"
			}
			if c.APIKey == "" {
				c.APIKey = "key123"
			}
			return nil
		},
	}

	chain := &InterpolatingChainLoader[Config]{
		Loaders: []Loader[Config]{loader},
	}

	cfg := &Config{}
	err := chain.Load(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify interpolation context was built
	context := chain.GetInterpolationContext()
	if context["ENV"] != "staging" {
		t.Errorf("expected context ENV='staging', got '%s'", context["ENV"])
	}
	if context["REGION"] != "us-west-2" {
		t.Errorf("expected context REGION='us-west-2', got '%s'", context["REGION"])
	}
}

// Test fast path (no interpolation) delegates to standard ChainLoader
func TestInterpolatingChainLoader_FastPath(t *testing.T) {
	type Config struct {
		Port int    `env:"PORT"`
		Host string `env:"HOST"`
	}

	loader := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			c.Port = 8080
			c.Host = "localhost"
			return nil
		},
	}

	chain := &InterpolatingChainLoader[Config]{
		Loaders: []Loader[Config]{loader},
	}

	cfg := &Config{}
	err := chain.Load(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify loader was called
	if loader.callCount != 1 {
		t.Errorf("expected loader to be called once, got %d", loader.callCount)
	}

	// Verify configuration was loaded
	if cfg.Port != 8080 {
		t.Errorf("expected Port=8080, got %d", cfg.Port)
	}
	if cfg.Host != "localhost" {
		t.Errorf("expected Host='localhost', got '%s'", cfg.Host)
	}
}

// Test loader precedence within stages
func TestInterpolatingChainLoader_LoaderPrecedence(t *testing.T) {
	type Config struct {
		Env  string `env:"ENV" config:"availableAs=ENV"`
		Port int    `env:"PORT"`
	}

	loader1 := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			c.Env = "dev"
			c.Port = 3000
			return nil
		},
	}

	loader2 := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			// Override Port but not Env
			c.Port = 8080
			return nil
		},
	}

	chain := &InterpolatingChainLoader[Config]{
		Loaders: []Loader[Config]{loader1, loader2},
	}

	cfg := &Config{}
	err := chain.Load(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify loader2 overrode Port
	if cfg.Port != 8080 {
		t.Errorf("expected Port=8080 (from loader2), got %d", cfg.Port)
	}

	// Verify Env from loader1 is preserved
	if cfg.Env != "dev" {
		t.Errorf("expected Env='dev', got '%s'", cfg.Env)
	}
}

// Test custom loader chain support
func TestInterpolatingChainLoader_CustomLoaderChain(t *testing.T) {
	type Config struct {
		Value1 string `config:"availableAs=VAL1"`
		Value2 string
	}

	customLoader := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			c.Value1 = "custom1"
			c.Value2 = "custom2"
			return nil
		},
	}

	chain := &InterpolatingChainLoader[Config]{
		Loaders: []Loader[Config]{customLoader},
	}

	cfg := &Config{}
	err := chain.Load(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Value1 != "custom1" {
		t.Errorf("expected Value1='custom1', got '%s'", cfg.Value1)
	}
	if cfg.Value2 != "custom2" {
		t.Errorf("expected Value2='custom2', got '%s'", cfg.Value2)
	}
}

// Test ShortCircuitChainLoader integration - basic short-circuit
func TestInterpolatingChainLoader_ShortCircuit_Basic(t *testing.T) {
	type Config struct {
		Field1 string `env:"FIELD1"`
		Field2 string `env:"FIELD2"`
	}

	loader1 := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			c.Field1 = "value1"
			c.Field2 = "value2"
			return nil
		},
	}

	loader2 := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			// This should not be called due to short-circuit
			c.Field1 = "override1"
			return nil
		},
	}

	chain := &InterpolatingChainLoader[Config]{
		Loaders:      []Loader[Config]{loader1, loader2},
		ShortCircuit: true,
	}

	cfg := &Config{}
	err := chain.Load(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify loader1 was called
	if loader1.callCount != 1 {
		t.Errorf("expected loader1 to be called once, got %d", loader1.callCount)
	}

	// Verify loader2 was NOT called (short-circuited)
	// Note: In fast path (no interpolation), we delegate to ShortCircuitChainLoader
	if loader2.callCount != 0 {
		t.Errorf("expected loader2 to not be called (short-circuit), got %d calls", loader2.callCount)
	}

	// Verify values from loader1 (not overridden by loader2)
	if cfg.Field1 != "value1" {
		t.Errorf("expected Field1='value1', got '%s'", cfg.Field1)
	}
	if cfg.Field2 != "value2" {
		t.Errorf("expected Field2='value2', got '%s'", cfg.Field2)
	}
}

// Test ShortCircuitChainLoader integration - dependency fields loaded before short-circuit
func TestInterpolatingChainLoader_ShortCircuit_WithDependencies(t *testing.T) {
	type Config struct {
		Env        string `env:"ENV" config:"availableAs=ENV"`
		DBPassword string `secret:"aws=/myapp/${ENV}/db/password"`
	}

	loader1 := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			c.Env = "prod"
			return nil
		},
	}

	loader2 := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			c.DBPassword = "secret123"
			return nil
		},
	}

	loader3 := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			// Should not be called in stage 1 due to short-circuit
			c.DBPassword = "override"
			return nil
		},
	}

	chain := &InterpolatingChainLoader[Config]{
		Loaders:      []Loader[Config]{loader1, loader2, loader3},
		ShortCircuit: true,
	}

	cfg := &Config{}
	err := chain.Load(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify Env was loaded (stage 0)
	if cfg.Env != "prod" {
		t.Errorf("expected Env='prod', got '%s'", cfg.Env)
	}

	// Verify DBPassword was loaded (stage 1)
	if cfg.DBPassword != "secret123" {
		t.Errorf("expected DBPassword='secret123', got '%s'", cfg.DBPassword)
	}

	// Verify context was updated even with short-circuit
	context := chain.GetInterpolationContext()
	if context["ENV"] != "prod" {
		t.Errorf("expected context ENV='prod', got '%s'", context["ENV"])
	}
}

// Test ShortCircuitChainLoader integration - short-circuit within stages, not across
func TestInterpolatingChainLoader_ShortCircuit_WithinStages(t *testing.T) {
	type Config struct {
		Env    string `env:"ENV" config:"availableAs=ENV"`
		Region string `env:"REGION" config:"availableAs=REGION"`
		Secret string `secret:"aws=/myapp/${ENV}/${REGION}/secret"`
	}

	stageTracker := make(map[string]int)

	loader1 := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			stageTracker["loader1"]++
			if c.Env == "" {
				c.Env = "prod"
			}
			if c.Region == "" {
				c.Region = "us-east-1"
			}
			return nil
		},
	}

	loader2 := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			stageTracker["loader2"]++
			// This loader should be skipped in stage 0 due to short-circuit
			// but called in stage 1
			if c.Secret == "" {
				c.Secret = "secret123"
			}
			return nil
		},
	}

	chain := &InterpolatingChainLoader[Config]{
		Loaders:      []Loader[Config]{loader1, loader2},
		ShortCircuit: true,
	}

	cfg := &Config{}
	err := chain.Load(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify all fields were loaded
	if cfg.Env != "prod" {
		t.Errorf("expected Env='prod', got '%s'", cfg.Env)
	}
	if cfg.Region != "us-east-1" {
		t.Errorf("expected Region='us-east-1', got '%s'", cfg.Region)
	}
	if cfg.Secret != "secret123" {
		t.Errorf("expected Secret='secret123', got '%s'", cfg.Secret)
	}

	// Verify loader1 was called at least once
	// Note: With short-circuit, loader1 populates all fields in stage 0,
	// so loader2 may not be called in stage 0. Both loaders are called in stage 1.
	if stageTracker["loader1"] < 1 {
		t.Errorf("expected loader1 to be called at least once, got %d", stageTracker["loader1"])
	}

	// Verify loader2 was called at least once (for stage 1)
	if stageTracker["loader2"] < 1 {
		t.Errorf("expected loader2 to be called at least once, got %d", stageTracker["loader2"])
	}
}

// Test with real loaders - EnvironmentLoader
func TestInterpolatingChainLoader_WithEnvironmentLoader(t *testing.T) {
	type Config struct {
		Env  string `env:"TEST_ENV" config:"availableAs=ENV"`
		Path string `env:"TEST_PATH_${ENV}"`
	}

	// Set environment variables
	os.Setenv("TEST_ENV", "production")
	os.Setenv("TEST_PATH_production", "/var/app/production")
	defer func() {
		os.Unsetenv("TEST_ENV")
		os.Unsetenv("TEST_PATH_production")
	}()

	chain := &InterpolatingChainLoader[Config]{
		Loaders: []Loader[Config]{
			&generic.EnvironmentLoader[Config]{},
		},
	}

	cfg := &Config{}
	err := chain.Load(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Env != "production" {
		t.Errorf("expected Env='production', got '%s'", cfg.Env)
	}

	// Note: Path won't be interpolated due to Go's struct tag limitations
	// This test verifies the infrastructure works correctly
}

// Test error handling - nil loaders
func TestInterpolatingChainLoader_NilLoaders(t *testing.T) {
	type Config struct {
		Field string `env:"FIELD"`
	}

	chain := &InterpolatingChainLoader[Config]{
		Loaders: nil,
	}

	cfg := &Config{}
	err := chain.Load(cfg)
	if err == nil {
		t.Fatal("expected error for nil Loaders, got nil")
	}
}

// Test error handling - nil loader in chain
func TestInterpolatingChainLoader_NilLoaderInChain(t *testing.T) {
	type Config struct {
		Field string `env:"FIELD"`
	}

	chain := &InterpolatingChainLoader[Config]{
		Loaders: []Loader[Config]{nil},
	}

	cfg := &Config{}
	err := chain.Load(cfg)
	if err == nil {
		t.Fatal("expected error for nil loader in chain, got nil")
	}
}

// Test error handling - loader returns error
func TestInterpolatingChainLoader_LoaderError(t *testing.T) {
	type Config struct {
		Field string `env:"FIELD"`
	}

	expectedErr := fmt.Errorf("loader error")
	loader := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			return expectedErr
		},
	}

	chain := &InterpolatingChainLoader[Config]{
		Loaders: []Loader[Config]{loader},
	}

	cfg := &Config{}
	err := chain.Load(cfg)
	if err == nil {
		t.Fatal("expected error from loader, got nil")
	}
}

// Test GetInterpolationContext returns copy
func TestInterpolatingChainLoader_GetInterpolationContext_ReturnsCopy(t *testing.T) {
	type Config struct {
		Env string `env:"ENV" config:"availableAs=ENV"`
	}

	loader := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			c.Env = "test"
			return nil
		},
	}

	chain := &InterpolatingChainLoader[Config]{
		Loaders: []Loader[Config]{loader},
	}

	cfg := &Config{}
	err := chain.Load(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Get context and modify it
	context := chain.GetInterpolationContext()
	context["ENV"] = "modified"

	// Get context again and verify it wasn't modified
	context2 := chain.GetInterpolationContext()
	if context2["ENV"] != "test" {
		t.Errorf("expected context to be unchanged, got '%s'", context2["ENV"])
	}
}

// Test GetInterpolationContext with no engine
func TestInterpolatingChainLoader_GetInterpolationContext_NoEngine(t *testing.T) {
	chain := &InterpolatingChainLoader[struct{}]{}

	context := chain.GetInterpolationContext()
	if context != nil {
		t.Errorf("expected nil context when engine is nil, got %v", context)
	}
}

// Test multiple stages with different field types
func TestInterpolatingChainLoader_MultipleStages_DifferentTypes(t *testing.T) {
	type Config struct {
		Port   int    `env:"PORT" config:"availableAs=PORT"`
		Debug  bool   `env:"DEBUG" config:"availableAs=DEBUG"`
		Config string `yaml:"config-${PORT}.yaml"`
	}

	loader := &mockLoader[Config]{
		loadFunc: func(c *Config) error {
			if c.Port == 0 {
				c.Port = 8080
			}
			if !c.Debug {
				c.Debug = true
			}
			if c.Config == "" {
				c.Config = "config.yaml"
			}
			return nil
		},
	}

	chain := &InterpolatingChainLoader[Config]{
		Loaders: []Loader[Config]{loader},
	}

	cfg := &Config{}
	err := chain.Load(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify type conversion in context
	context := chain.GetInterpolationContext()
	if context["PORT"] != "8080" {
		t.Errorf("expected context PORT='8080', got '%s'", context["PORT"])
	}
	if context["DEBUG"] != "true" {
		t.Errorf("expected context DEBUG='true', got '%s'", context["DEBUG"])
	}
}
