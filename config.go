package config

import (
	"os"
	"reflect"

	"github.com/crazywolf132/secretfetch"
	"github.com/go-playground/validator/v10"
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

type Loader[T any] interface {
	Load(c *T) error
}

func DefaultConfigLoaders[T any]() []Loader[T] {
	return []Loader[T]{
		&EnvironmentLoader[T]{},
		&CommandLineLoader[T]{Args: os.Args[1:]},
	}
}

func isConfigFullyPopulated[T any](c *T) bool {
	if c == nil {
		return false
	}
	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)
		if structField.PkgPath != "" { // skip unexported
			continue
		}
		if isZero(field) {
			return false
		}
	}
	return true
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Map, reflect.Slice, reflect.Ptr:
		return v.IsNil()
	case reflect.Interface:
		if v.IsNil() {
			return true
		}
		// If the underlying value is a nil pointer, treat as zero
		underlying := v.Elem()
		if (underlying.Kind() == reflect.Ptr || underlying.Kind() == reflect.Interface) && underlying.IsNil() {
			return true
		}
		// For other types, recursively check if the underlying value is zero
		return isZero(underlying)
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZero(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == 0
	case reflect.String:
		return v.String() == ""
	}
	return false
}
