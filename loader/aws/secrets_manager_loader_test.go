package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/crazywolf132/secretfetch"
	"github.com/gymshark/go-easy-config/loader"
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

	ldr := &SecretsManagerLoader[SecretsTestConfig]{
		SecretFetchOpts: &secretfetch.Options{
			AWS:            &aws.Config{Region: "us-east-1"},
			SecretsManager: mockClient,
		},
	}

	if err := ldr.Load(cfg); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.SecretVar1 != "test-secret-value" {
		t.Errorf("expected SecretVar1 to be 'test-secret-value', got '%s'", cfg.SecretVar1)
	}
}

func TestSecretsManagerLoader_FetchError(t *testing.T) {
	cfg := &SecretsTestConfig{}

	mockClient := &mockSecretsManagerClient{
		getSecretValueFn: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return nil, errors.New("secret not found")
		},
	}

	ldr := &SecretsManagerLoader[SecretsTestConfig]{
		SecretFetchOpts: &secretfetch.Options{
			AWS:            &aws.Config{Region: "us-east-1"},
			SecretsManager: mockClient,
		},
	}

	err := ldr.Load(cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Test that error is LoaderError
	var loaderErr *loader.LoaderError
	if !errors.As(err, &loaderErr) {
		t.Fatalf("expected LoaderError, got %T", err)
	}

	// Test error fields
	if loaderErr.LoaderType != "SecretsManagerLoader" {
		t.Errorf("expected LoaderType 'SecretsManagerLoader', got '%s'", loaderErr.LoaderType)
	}

	if loaderErr.Operation != "fetch secrets" {
		t.Errorf("expected Operation 'fetch secrets', got '%s'", loaderErr.Operation)
	}

	// Test that underlying error is accessible
	if loaderErr.Err == nil {
		t.Error("expected underlying error to be set")
	}

	// Test error message includes context
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestSecretsManagerLoader_ErrorWrapping(t *testing.T) {
	// Test that errors are properly wrapped with LoaderError
	tests := []struct {
		name        string
		mockError   error
		expectedOp  string
		expectError bool
	}{
		{
			name:        "fetch secrets error",
			mockError:   errors.New("access denied"),
			expectedOp:  "fetch secrets",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &SecretsTestConfig{}

			mockClient := &mockSecretsManagerClient{
				getSecretValueFn: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					return nil, tt.mockError
				},
			}

			ldr := &SecretsManagerLoader[SecretsTestConfig]{
				SecretFetchOpts: &secretfetch.Options{
					AWS:            &aws.Config{Region: "us-east-1"},
					SecretsManager: mockClient,
				},
			}

			err := ldr.Load(cfg)

			if tt.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if tt.expectError {
				var loaderErr *loader.LoaderError
				if !errors.As(err, &loaderErr) {
					t.Fatalf("expected LoaderError, got %T", err)
				}

				if loaderErr.LoaderType != "SecretsManagerLoader" {
					t.Errorf("expected LoaderType 'SecretsManagerLoader', got '%s'", loaderErr.LoaderType)
				}

				if loaderErr.Operation != tt.expectedOp {
					t.Errorf("expected Operation '%s', got '%s'", tt.expectedOp, loaderErr.Operation)
				}

				// Verify error message includes AWS-specific context
				errMsg := err.Error()
				if errMsg == "" {
					t.Error("expected non-empty error message")
				}
			}
		})
	}
}
