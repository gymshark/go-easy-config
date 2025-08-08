package generic

import (
	"testing"
)

type CmdTestConfig struct {
	CmdVar1 string `clap:"--cmdvar1"`
}

func TestCommandLineLoader_Load(t *testing.T) {
	cfg := &CmdTestConfig{}
	args := []string{"--cmdvar1", "cmd_value"}
	loader := &CommandLineLoader[CmdTestConfig]{Args: args}
	if err := loader.Load(cfg); err != nil {
		t.Fatalf("CommandLineLoader failed: %v", err)
	}
	if cfg.CmdVar1 != "cmd_value" {
		t.Errorf("CmdVar1 not loaded, got: %s", cfg.CmdVar1)
	}
}
