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

type Option[C any] func(*Handler[C])

type Handler[C any] struct {
	Validator   *validator.Validate
	Loaders     []Loader[C]
	chainLoader *ChainLoader[C]
}

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
	handler.chainLoader = &ChainLoader[C]{Loaders: handler.Loaders}
	return handler
}

func WithValidator[C any](v *validator.Validate) Option[C] {
	return func(h *Handler[C]) {
		if v == nil {
			v = DefaultConfigValidator()
		}
		h.Validator = v
		// Ensure chainLoader is up to date
		h.chainLoader = &ChainLoader[C]{Loaders: h.Loaders}
	}
}

func WithLoaders[C any](loaders ...Loader[C]) Option[C] {
	return func(h *Handler[C]) {
		h.Loaders = loaders
		// Ensure chainLoader is up to date
		h.chainLoader = &ChainLoader[C]{Loaders: h.Loaders}
	}
}

func (c *Handler[C]) Load(cfg *C) error {
	return c.chainLoader.Load(cfg)
}

func (c *Handler[C]) Validate(cfg *C) error {
	return c.Validator.Struct(cfg)
}

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
