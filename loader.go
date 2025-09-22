package config

// Loader defines the interface for configuration loaders.
// Each loader is responsible for populating configuration from a specific source.
type Loader[T any] interface {
	// Load populates the configuration struct from the loader's source.
	// It should not overwrite existing non-zero values unless explicitly designed to do so.
	Load(c *T) error
}
