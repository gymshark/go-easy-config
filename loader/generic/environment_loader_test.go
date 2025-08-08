package generic

import (
	"os"
	"testing"
)

type EnvTestConfig struct {
	EnvVar1 string `env:"TEST_ENV_VAR1"`
}

func TestEnvironmentLoader_Load(t *testing.T) {
	os.Setenv("TEST_ENV_VAR1", "env_value")
	cfg := &EnvTestConfig{}
	loader := &EnvironmentLoader[EnvTestConfig]{}
	if err := loader.Load(cfg); err != nil {
		t.Fatalf("EnvironmentLoader failed: %v", err)
	}
	if cfg.EnvVar1 != "env_value" {
		t.Errorf("EnvVar1 not loaded, got: %s", cfg.EnvVar1)
	}
}
