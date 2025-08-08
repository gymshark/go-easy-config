package config

import (
	"os"
	"testing"
)

type ShortCircuitTestConfig struct {
	EnvVar1    string `env:"TEST_ENV_VAR1"`
	CmdVar1    string `clap:"--cmdvar1"`
	SecretVar1 string
}

func TestShortCircuitChainLoader_Load(t *testing.T) {
	os.Setenv("TEST_ENV_VAR1", "env_value")
	cfg := &ShortCircuitTestConfig{SecretVar1: "secret_value"}
	loaders := []Loader[ShortCircuitTestConfig]{
		&EnvironmentLoader[ShortCircuitTestConfig]{},
		&CommandLineLoader[ShortCircuitTestConfig]{Args: []string{"--cmdvar1", "cmd_value", "--envvar1", "env_value"}},
	}
	chain := &ShortCircuitChainLoader[ShortCircuitTestConfig]{Loaders: loaders}
	if err := chain.Load(cfg); err != nil {
		t.Fatalf("ShortCircuitChainLoader failed: %v", err)
	}
	if cfg.EnvVar1 != "env_value" {
		t.Errorf("EnvVar1 not loaded, got: %s", cfg.EnvVar1)
	}
	if cfg.CmdVar1 != "cmd_value" {
		t.Errorf("CmdVar1 not loaded, got: %s", cfg.CmdVar1)
	}
}
