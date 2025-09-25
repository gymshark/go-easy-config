// Package aws provides loaders for AWS-specific configuration sources.
package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/crazywolf132/secretfetch"
)

// SecretsManagerLoader loads configuration values from AWS Secrets Manager.
// It uses the secretfetch library to handle secret retrieval and supports
// fields tagged with `secret:"aws=secret-name"`.
// Unlike secretfetch directly, this loader can handle structs with mixed tag types
// by only processing fields that have secret tags.
type SecretsManagerLoader[T any] struct {
	SecretFetchOpts *secretfetch.Options
}

// Load fetches secrets from AWS Secrets Manager for fields with appropriate tags.
// It handles mixed tag scenarios by only processing fields with secret tags.
func (s *SecretsManagerLoader[T]) Load(c *T) error {
	opts := s.SecretFetchOpts
	if opts == nil {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			log.Fatalf("error loading AWS config: %v", err)
		}
		
		opts = &secretfetch.Options{
			AWS: &cfg,
		}
	}

	// Check if any fields have secret tags before calling secretfetch
	if !hasSecretTags(c) {
		return nil // No secret fields to process
	}

	// Create a temporary struct with only secret-tagged fields
	tempStruct, fieldMap, err := createSecretOnlyStruct(c)
	if err != nil {
		return fmt.Errorf("error creating secret-only struct: %w", err)
	}

	// Fetch secrets into the temporary struct
	if err := secretfetch.Fetch(context.Background(), tempStruct, opts); err != nil {
		return fmt.Errorf("error fetching secrets: %w", err)
	}

	// Copy values back to the original struct
	return copySecretValues(c, tempStruct, fieldMap)
}
