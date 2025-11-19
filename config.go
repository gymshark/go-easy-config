// Package config provides a lightweight configuration loader and validator for Go applications.
// It supports loading configuration from multiple sources including environment variables,
// command-line flags, and AWS Secrets Manager.
package config

import (
	"os"

	"github.com/crazywolf132/secretfetch"
	"github.com/go-playground/validator/v10"
	"github.com/gymshark/go-easy-config/loader/generic"
)

var (
	defaultSecretFetchOpts = &secretfetch.Options{}
)

// Option is a functional option for configuring a Handler.
type Option[C any] func(*Handler[C])

// Handler manages configuration loading and validation for a specific configuration type.
type Handler[C any] struct {
	Validator   *validator.Validate
	Loaders     []Loader[C]
	chainLoader *InterpolatingChainLoader[C] // Internal chain loader with interpolation support
}

// NewConfigHandler creates a new configuration handler with default loaders and validator.
// Default loaders include environment variables and command-line arguments.
// The handler uses InterpolatingChainLoader which automatically detects and handles
// variable interpolation in struct tags while maintaining full backward compatibility.
func NewConfigHandler[C any](options ...Option[C]) *Handler[C] {
	loaders := DefaultConfigLoaders[C]()
	handler := &Handler[C]{
		Validator: DefaultConfigValidator(),
		Loaders:   loaders,
	}
	if options != nil {
		for _, opt := range options {
			opt(handler)
		}
	}
	handler.chainLoader = &InterpolatingChainLoader[C]{Loaders: handler.Loaders}
	return handler
}

// WithValidator sets a custom validator for the configuration handler.
func WithValidator[C any](v *validator.Validate) Option[C] {
	return func(h *Handler[C]) {
		if v == nil {
			v = DefaultConfigValidator()
		}
		h.Validator = v
		// Ensure chainLoader is up to date
		h.chainLoader = &InterpolatingChainLoader[C]{Loaders: h.Loaders}
	}
}

// WithLoaders sets custom loaders for the configuration handler.
// Loaders are executed in the order provided.
// The InterpolatingChainLoader automatically handles variable interpolation
// while maintaining the specified loader precedence.
func WithLoaders[C any](loaders ...Loader[C]) Option[C] {
	return func(h *Handler[C]) {
		h.Loaders = loaders
		// Ensure chainLoader is up to date
		h.chainLoader = &InterpolatingChainLoader[C]{Loaders: h.Loaders}
	}
}

// Load populates the configuration struct using all configured loaders in sequence.
func (c *Handler[C]) Load(cfg *C) error {
	return c.chainLoader.Load(cfg)
}

// Validate validates the configuration struct using the configured validator.
// Returns ValidationError wrapping any validator errors for consistent error handling.
func (c *Handler[C]) Validate(cfg *C) error {
	err := c.Validator.Struct(cfg)
	if err != nil {
		// Wrap validator error in ValidationError for consistency
		return &ValidationError{
			FieldName: "<multiple>",
			Rule:      "<multiple>",
			Err:       err,
		}
	}
	return nil
}

// LoadAndValidate loads and then validates the configuration in a single operation.
func (c *Handler[C]) LoadAndValidate(cfg *C) error {
	err := c.Load(cfg)
	if err != nil {
		return err
	}

	err = c.Validate(cfg)
	if err != nil {
		return err
	}

	return nil
}

func DefaultConfigValidator() *validator.Validate {
	defaultValidator := NewValidator()
	return &defaultValidator
}

func DefaultConfigLoaders[T any]() []Loader[T] {
	return []Loader[T]{
		&generic.EnvironmentLoader[T]{},
		&generic.CommandLineLoader[T]{Args: os.Args[1:]},
	}
}
