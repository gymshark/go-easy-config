package config

import (
	"context"
	"fmt"
	"github.com/crazywolf132/secretfetch"
)

type SecretsManagerLoader[T any] struct {
	SecretFetchOpts *secretfetch.Options
}

func (s *SecretsManagerLoader[T]) Load(c *T) error {
	opts := s.SecretFetchOpts
	if opts == nil {
		opts = &secretfetch.Options{}
	}
	if err := secretfetch.Fetch(context.Background(), c, opts); err != nil {
		return fmt.Errorf("error fetching secrets: %w", err)
	}
	return nil
}
