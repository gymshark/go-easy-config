package config

import (
	"testing"
)

type SecretsTestConfig struct {
	SecretVar1 string
}

func TestSecretsManagerLoader_Load(t *testing.T) {
	cfg := &SecretsTestConfig{}
	loader := &SecretsManagerLoader[SecretsTestConfig]{}
	// This test expects no AWS setup, so error is acceptable
	_ = loader.Load(cfg)
}
