package generic

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type EnvironmentLoader[T any] struct{}

func (e *EnvironmentLoader[T]) Load(c *T) error {
	if err := env.Parse(c); err != nil {
		return fmt.Errorf("error parsing environment variables: %w", err)
	}
	return nil
}
