package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/crazywolf132/secretfetch"
)

type mockSecretsManagerClient struct {
	getSecretValueFn func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

func (m *mockSecretsManagerClient) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return m.getSecretValueFn(ctx, params, optFns...)
}

type SecretsTestConfig struct {
	SecretVar1 string `secret:"aws=test-secret"`
}

func TestSecretsManagerLoader_Load(t *testing.T) {
	cfg := &SecretsTestConfig{}

	mockClient := &mockSecretsManagerClient{
		getSecretValueFn: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String("test-secret-value"),
			}, nil
		},
	}

	loader := &SecretsManagerLoader[SecretsTestConfig]{
		SecretFetchOpts: &secretfetch.Options{
			AWS:            &aws.Config{Region: "us-east-1"},
			SecretsManager: mockClient,
		},
	}

	if err := loader.Load(cfg); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.SecretVar1 != "test-secret-value" {
		t.Errorf("expected SecretVar1 to be 'test-secret-value', got '%s'", cfg.SecretVar1)
	}
}
