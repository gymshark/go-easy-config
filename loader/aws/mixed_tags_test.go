package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/crazywolf132/secretfetch"
)

// MixedTagsConfig demonstrates the issue where some fields have secret tags
// and others have different tags (env, clap, etc.)
type MixedTagsConfig struct {
	SecretValue string `secret:"aws=test-secret"`
	EnvValue    string `env:"TEST_ENV"`
	Port        int    `env:"PORT" clap:"port"`
}

func TestSecretsManagerLoader_MixedTags(t *testing.T) {
	cfg := &MixedTagsConfig{}

	mockClient := &mockSecretsManagerClient{
		getSecretValueFn: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String("secret-from-aws"),
			}, nil
		},
	}

	loader := &SecretsManagerLoader[MixedTagsConfig]{
		SecretFetchOpts: &secretfetch.Options{
			AWS:            &aws.Config{Region: "us-east-1"},
			SecretsManager: mockClient,
		},
	}

	// This should now succeed with the mixed tags handling
	err := loader.Load(cfg)
	if err != nil {
		t.Fatalf("expected no error with mixed tags handling, got: %v", err)
	}

	// Verify the secret field was populated
	if cfg.SecretValue != "secret-from-aws" {
		t.Errorf("expected SecretValue to be 'secret-from-aws', got '%s'", cfg.SecretValue)
	}

	// Verify other fields remain unchanged (they should be empty since we didn't set them)
	if cfg.EnvValue != "" {
		t.Errorf("expected EnvValue to remain empty, got '%s'", cfg.EnvValue)
	}
	if cfg.Port != 0 {
		t.Errorf("expected Port to remain 0, got %d", cfg.Port)
	}
}

// NoSecretTagsConfig has no secret tags at all
type NoSecretTagsConfig struct {
	EnvValue string `env:"TEST_ENV"`
	Port     int    `env:"PORT" clap:"port"`
}

func TestSecretsManagerLoader_NoSecretTags(t *testing.T) {
	cfg := &NoSecretTagsConfig{}

	mockClient := &mockSecretsManagerClient{
		getSecretValueFn: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			t.Error("mock should not be called when there are no secret tags")
			return nil, nil
		},
	}

	loader := &SecretsManagerLoader[NoSecretTagsConfig]{
		SecretFetchOpts: &secretfetch.Options{
			AWS:            &aws.Config{Region: "us-east-1"},
			SecretsManager: mockClient,
		},
	}

	// This should succeed and do nothing
	err := loader.Load(cfg)
	if err != nil {
		t.Fatalf("expected no error when no secret tags present, got: %v", err)
	}

	// Verify fields remain unchanged
	if cfg.EnvValue != "" {
		t.Errorf("expected EnvValue to remain empty, got '%s'", cfg.EnvValue)
	}
	if cfg.Port != 0 {
		t.Errorf("expected Port to remain 0, got %d", cfg.Port)
	}
}
