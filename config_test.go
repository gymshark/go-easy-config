package config

import (
	"os"
	"reflect"
	"testing"
)

type TestConfig struct {
	EnvVar1    string `env:"TEST_ENV_VAR1"`
	SecretVar1 string
	CmdVar1    string `clap:"--cmdvar1"`
}

const (
	CmdValue = "cmd_value"
	EnvValue = "env_value"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{SecretVar1: "secret_value"}
	config := NewConfigHandler[TestConfig]()
	if err := config.Load(cfg); err != nil {
		t.Fatalf("ConfigHandler.Load failed: %v", err)
	}
	if cfg.EnvVar1 != EnvValue {
		t.Errorf("EnvVar1 not loaded, got: %s", cfg.EnvVar1)
	}
}

func TestValidateConfig(t *testing.T) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{EnvVar1: EnvValue}
	config := NewConfigHandler[TestConfig]()
	if err := config.Validate(cfg); err != nil {
		t.Fatalf("ConfigHandler.Validate failed: %v", err)
	}
}

func TestLoadAndValidateConfig(t *testing.T) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{SecretVar1: "secret_value"}
	config := NewConfigHandler[TestConfig]()
	if err := config.LoadAndValidate(cfg); err != nil {
		t.Fatalf("ConfigHandler.LoadAndValidate failed: %v", err)
	}
	if cfg.EnvVar1 != EnvValue {
		t.Errorf("EnvVar1 not loaded, got: %s", cfg.EnvVar1)
	}
}

func TestWithValidator(t *testing.T) {
	customValidator := DefaultConfigValidator()
	handler := NewConfigHandler[TestConfig](WithValidator[TestConfig](customValidator))
	if handler.Validator != customValidator {
		t.Errorf("WithValidator did not set custom validator")
	}
}

func TestWithLoaders(t *testing.T) {
	customLoader := &EnvironmentLoader[TestConfig]{}
	handler := NewConfigHandler[TestConfig](WithLoaders(customLoader))
	if len(handler.Loaders) != 1 {
		t.Errorf("WithLoaders did not set custom loaders")
	}
	if _, ok := handler.Loaders[0].(*EnvironmentLoader[TestConfig]); !ok {
		t.Errorf("WithLoaders did not set correct loader type")
	}
}

func TestIsZero_AllTypes(t *testing.T) {
	var (
		str                    = ""
		b                      = false
		intVal     int64       = 0
		uintVal    uint64      = 0
		floatVal   float64     = 0
		complexVal complex128  = 0
		ptr        *int        = nil
		iface      interface{} = nil
		arr        [2]int
		ch         chan int
		m          map[string]int
		slice      []int
	)
	cases := []struct {
		val  reflect.Value
		want bool
	}{
		{reflect.ValueOf(str), true},
		{reflect.ValueOf(b), true},
		{reflect.ValueOf(intVal), true},
		{reflect.ValueOf(uintVal), true},
		{reflect.ValueOf(floatVal), true},
		{reflect.ValueOf(complexVal), true},
		{reflect.ValueOf(ptr), true},
		{reflect.ValueOf(iface), false}, // interface is nil, expect false due to Go reflect semantics
		{reflect.ValueOf(arr), true},    // array of zero values, expect true
		{reflect.ValueOf(ch), true},
		{reflect.ValueOf(m), true},
		{reflect.ValueOf(slice), true},
	}
	for i, c := range cases {
		if got := isZero(c.val); got != c.want {
			t.Errorf("isZero case %d failed: got %v, want %v", i, got, c.want)
		}
	}
}

func BenchmarkEnvironmentLoader_Load(b *testing.B) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{}
	loader := &EnvironmentLoader[TestConfig]{}
	for i := 0; i < b.N; i++ {
		_ = loader.Load(cfg)
	}
}

func BenchmarkCommandLineLoader_Load(b *testing.B) {
	cfg := &TestConfig{}
	args := []string{"--cmdvar1", CmdValue}
	loader := &CommandLineLoader[TestConfig]{Args: args}
	for i := 0; i < b.N; i++ {
		_ = loader.Load(cfg)
	}
}

func BenchmarkChainLoader_Load(b *testing.B) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{}
	loaders := []Loader[TestConfig]{
		&EnvironmentLoader[TestConfig]{},
		&CommandLineLoader[TestConfig]{Args: []string{"--cmdvar1", CmdValue}},
	}
	chain := &ChainLoader[TestConfig]{Loaders: loaders}
	for i := 0; i < b.N; i++ {
		_ = chain.Load(cfg)
	}
}

func BenchmarkConfigHandler_Load(b *testing.B) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{}
	config := NewConfigHandler[TestConfig]()
	for i := 0; i < b.N; i++ {
		_ = config.Load(cfg)
	}
}

func BenchmarkConfigHandler_LoadAndValidate(b *testing.B) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{}
	config := NewConfigHandler[TestConfig]()
	for i := 0; i < b.N; i++ {
		_ = config.LoadAndValidate(cfg)
	}
}

func BenchmarkShortCircuitChainLoader_Load(b *testing.B) {
	os.Setenv("TEST_ENV_VAR1", EnvValue)
	cfg := &TestConfig{}
	loaders := []Loader[TestConfig]{
		&EnvironmentLoader[TestConfig]{},
		&CommandLineLoader[TestConfig]{Args: []string{"--cmdvar1", CmdValue}},
	}
	chain := &ShortCircuitChainLoader[TestConfig]{Loaders: loaders}
	for i := 0; i < b.N; i++ {
		_ = chain.Load(cfg)
	}
}
