package config

import (
	"fmt"

	"github.com/gymshark/go-easy-config/utils"
)

// ShortCircuitChainLoader executes loaders in sequence but stops early
// when all exported fields in the configuration struct are populated.
// This can improve performance when not all loaders are needed.
type ShortCircuitChainLoader[T any] struct {
	Loaders []Loader[T]
}

// Load executes loaders until all exported fields are populated or all loaders are exhausted.
func (l *ShortCircuitChainLoader[T]) Load(c *T) error {
	if l.Loaders == nil {
		return fmt.Errorf("ShortCircuitChainLoader.Loaders is nil")
	}
	for i, loader := range l.Loaders {
		if loader == nil {
			return fmt.Errorf("ShortCircuitChainLoader loader at index %d is nil", i)
		}
		// Stop early if all fields are populated
		if utils.IsConfigFullyPopulated(c) {
			break
		}
		if err := loader.Load(c); err != nil {
			return fmt.Errorf("error loading config in loader at index %d: %w", i, err)
		}
	}
	return nil
}
