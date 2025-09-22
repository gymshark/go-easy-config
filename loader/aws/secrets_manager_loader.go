package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/crazywolf132/secretfetch"
)

type SecretsManagerLoader[T any] struct {
	SecretFetchOpts *secretfetch.Options
}

func (s *SecretsManagerLoader[T]) Load(c *T) error {
	opts := s.SecretFetchOpts
	if opts == nil {
		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "" {
			awsRegion = "us-east-1"
		}
		opts = &secretfetch.Options{
			AWS: &aws.Config{Region: awsRegion},
		}
	}
	if err := secretfetch.Fetch(context.Background(), c, opts); err != nil {
		return fmt.Errorf("error fetching secrets: %w", err)
	}
	return nil
}
