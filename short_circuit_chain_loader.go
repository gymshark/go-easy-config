package config

import (
	"fmt"
)

type ShortCircuitChainLoader[T any] struct {
	Loaders []Loader[T]
}

func (l *ShortCircuitChainLoader[T]) Load(c *T) error {
	if l.Loaders == nil {
		return fmt.Errorf("ShortCircuitChainLoader.Loaders is nil")
	}
	for i, loader := range l.Loaders {
		if loader == nil {
			return fmt.Errorf("ShortCircuitChainLoader loader at index %d is nil", i)
		}
		if isConfigFullyPopulated(c) {
			break
		}
		if err := loader.Load(c); err != nil {
			return fmt.Errorf("error loading config in loader at index %d: %w", i, err)
		}
	}
	return nil
}
