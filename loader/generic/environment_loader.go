// Package generic provides loaders for common configuration sources.
package generic

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// EnvironmentLoader loads configuration from environment variables.
// It supports fields tagged with `env:"VARIABLE_NAME"`.
type EnvironmentLoader[T any] struct{}

// Load populates configuration fields from environment variables.
func (e *EnvironmentLoader[T]) Load(c *T) error {
	if err := env.Parse(c); err != nil {
		return fmt.Errorf("error parsing environment variables: %w", err)
	}
	return nil
}
