package config

import (
	"os"
	"testing"
)

type ChainTestConfig struct {
	EnvVar1 string `env:"TEST_ENV_VAR1"`
	CmdVar1 string `clap:"--cmdvar1"`
}

func TestChainLoader_Load(t *testing.T) {
	os.Setenv("TEST_ENV_VAR1", "env_value")
	cfg := &ChainTestConfig{}
	loaders := []Loader[ChainTestConfig]{
		&EnvironmentLoader[ChainTestConfig]{},
		&CommandLineLoader[ChainTestConfig]{Args: []string{"--cmdvar1", "cmd_value"}},
	}
	chain := &ChainLoader[ChainTestConfig]{Loaders: loaders}
	if err := chain.Load(cfg); err != nil {
		t.Fatalf("ChainLoader failed: %v", err)
	}
	if cfg.EnvVar1 != "env_value" {
		t.Errorf("EnvVar1 not loaded, got: %s", cfg.EnvVar1)
	}
	if cfg.CmdVar1 != "cmd_value" {
		t.Errorf("CmdVar1 not loaded, got: %s", cfg.CmdVar1)
	}
}
