package aws

import (
	"github.com/ianlopshire/go-ssm-config"
)

// SSMParameterStoreLoader loads configuration from AWS Systems Manager Parameter Store.
// It uses the go-ssm-config library to fetch parameters based on struct tags.
type SSMParameterStoreLoader[T any] struct {
	Path string // Base path for parameter lookup in Parameter Store
}

// Load fetches parameters from SSM Parameter Store for fields with appropriate tags.
func (s *SSMParameterStoreLoader[T]) Load(c *T) error {
	return ssmconfig.Process(s.Path, c)
}
