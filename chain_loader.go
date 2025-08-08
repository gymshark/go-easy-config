package config

import (
	"fmt"
)

type ChainLoader[T any] struct {
	Loaders []Loader[T]
}

func (l *ChainLoader[T]) Load(c *T) error {
	if l.Loaders == nil {
		return fmt.Errorf("ChainLoader.Loaders is nil")
	}
	for i, loader := range l.Loaders {
		if loader == nil {
			return fmt.Errorf("ChainLoader loader at index %d is nil", i)
		}
		if err := loader.Load(c); err != nil {
			return fmt.Errorf("error loading config in loader at index %d: %w", i, err)
		}
	}
	return nil
}
