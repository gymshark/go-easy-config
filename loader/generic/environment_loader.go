// Package generic provides loaders for common configuration sources.
package generic

import (
	"github.com/caarlos0/env/v11"
	"github.com/gymshark/go-easy-config/loader"
)

// EnvironmentLoader loads configuration from environment variables.
// It supports fields tagged with `env:"VARIABLE_NAME"`.
type EnvironmentLoader[T any] struct{}

// Load populates configuration fields from environment variables.
func (e *EnvironmentLoader[T]) Load(c *T) error {
	if err := env.Parse(c); err != nil {
		return &loader.LoaderError{
			LoaderType: "EnvironmentLoader",
			Operation:  "parse environment variables",
			Err:        err,
		}
	}
	return nil
}
