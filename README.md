# go-easy-config

A lightweight configuration loader and validator for Go applications. Supports environment variables, command-line flags, and optional AWS Secrets Manager integration via secretfetch.

## Contents
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Define Your Configuration Struct](#define-your-configuration-struct)
  - [Load and Validate Configuration](#load-and-validate-configuration)
  - [AWS Secrets Manager Integration](#aws-secrets-manager-integration)
  - [Customising Loaders and Validators](#customising-loaders-and-validators)
  - [Types of Configuration Sources](#types-of-configuration-sources)
  - [Loader Order and Customisation](#loader-order-and-customisation)
    - [Custom Loader Order Example](#custom-loader-order-example)
    - [ShortCircuitChainLoader (Optional Short-Circuiting)](#shortcircuitchainloader-optional-short-circuiting)
    - [Providing Your Own Chain Loader](#providing-your-own-chain-loader)
    - [Providing Your Own Loader](#providing-your-own-loader)
- [Validation](#validation)
  - [Advanced Validation](#advanced-validation)
- [Testing](#testing)
- [License](#license)

## Features
- Load configuration from environment variables
- Parse command-line flags
- Fetch secrets from AWS Secrets Manager (optional)
- Validate configuration using go-playground/validator
- Modular loader design for extensibility

## Installation

```shell
go get github.com/gymshark/go-easy-config
```

## Usage

### Define Your Configuration Struct

```go
package main

import (
	"github.com/gymshark/go-easy-config"
)

type AppConfig struct {
	Port int    `env:"PORT" clap:"port" validate:"required,min=1,max=65535"`
	Env  string `env:"ENV" clap:"env" validate:"required,oneof=dev prod"`
	Secret string `env:"SECRET" clap:"secret"`
}
```

### Load and Validate Configuration

```go
func main() {
	handler := config.NewConfigHandler[AppConfig]()
	var cfg AppConfig
	if err := handler.Load(&cfg); err != nil {
		panic(err)
	}
	if err := handler.Validate(&cfg); err != nil {
		panic(err)
	}
	// Use cfg...
}
```

### AWS Secrets Manager Integration

To fetch secrets, add fields with the `secretfetch` tag and configure AWS credentials:

```go
type SecretsConfig struct {
	DBPassword string `secretfetch:"/prod/db/password"`
}
```

### Customising Loaders and Validators

You can provide custom loaders or validators:

```go
handler := config.NewConfigHandler[AppConfig](
	config.WithValidator(customValidator),
	config.WithLoaders(customLoader),
)
```

### Types of Configuration Sources

#### Environment Variables (`env` tag)
Fields tagged with `env:"NAME"` are loaded from environment variables using [caarlos0/env](https://github.com/caarlos0/env).

#### Command-Line Arguments (`clap` tag)
Fields tagged with `clap:"name"` are loaded from command-line flags using [go-clap](https://github.com/fred1268/go-clap).

#### AWS Secrets Manager (`secretfetch` tag)
Fields tagged with `secretfetch:"/path/to/secret"` are loaded from AWS Secrets Manager using [secretfetch](https://github.com/crazywolf132/secretfetch).

### Loader Order and Customisation
By default, configuration is loaded in the following order:
1. Environment variables
2. Command-line arguments

Each loader processes the configuration struct in sequence. If a variable is set by an earlier loader, it may be overridden by a later loader if the same field is present in multiple sources. For example, a value set via an environment variable will be replaced if a command-line argument for the same field is provided, and both may be replaced if a secret is fetched from AWS Secrets Manager for that field.

You can customise the loader order and add/remove loaders using `WithLoaders`.

#### Custom Loader Order Example

To specify your own loader order:

```go
import (
	"github.com/gymshark/go-easy-config"
	"os"
)

loaders := []config.Loader[AppConfig]{
	&config.CommandLineLoader[AppConfig]{Args: os.Args[1:]},
	&config.EnvironmentLoader[AppConfig]{},
	&config.SecretsManagerLoader[AppConfig]{},
}
handler := config.NewConfigHandler[AppConfig](
	config.WithLoaders(loaders...),
)
```

#### ShortCircuitChainLoader (Optional Short-Circuiting)

If you want to stop loading as soon as all exported fields in your configuration struct are populated, use the `ShortCircuitChainLoader`:

```go
loaders := []config.Loader[AppConfig]{
	&config.EnvironmentLoader[AppConfig]{},
	&config.CommandLineLoader[AppConfig]{Args: os.Args[1:]},
}
chain := &config.ShortCircuitChainLoader[AppConfig]{Loaders: loaders}
err := chain.Load(&cfg)
```

This loader will stop processing further loaders as soon as all exported fields are set, which can improve efficiency if not all sources are needed. The standard `ChainLoader` always runs all loaders in sequence.

#### Providing Your Own Chain Loader
A chain loader orchestrates multiple loaders, executing them in sequence. Each loader may override values set by previous loaders. The chain loader delegates to each loader in turn.

A loader is a single component that implements the `Loader` interface and loads configuration from a specific source.

Example chain loader:

```go
type CustomChainLoader[T any] struct {
	Loaders []config.Loader[T]
}

func (l *CustomChainLoader[T]) Load(c *T) error {
	for _, loader := range l.Loaders {
		if err := loader.Load(c); err != nil {
			return err
		}
	}
	return nil
}
```

Use your custom chain loader to control the order and logic of configuration loading, including conditional logic or error handling.

#### Providing Your Own Loader

Implement your own loader by satisfying the `Loader` interface:

```go
type Loader[T any] interface {
	Load(c *T) error
}
```

Example custom loader:

```go
type FileLoader[T any] struct {
	Path string
}

func (f *FileLoader[T]) Load(c *T) error {
	// Implement file reading and unmarshalling logic here
	return nil
}
```

Use your custom loader in the loader chain:

```go
loaders := []config.Loader[AppConfig]{
	&FileLoader[AppConfig]{Path: "config.yaml"},
	&config.EnvironmentLoader[AppConfig]{},
}
handler := config.NewConfigHandler[AppConfig](
	config.WithLoaders(loaders...),
)
```

## Validation

Configuration structs can be validated using [go-playground/validator](https://github.com/go-playground/validator). Use the `validate` struct tag to specify rules:

```go
Port int `validate:"required,min=1,max=65535"`
Env  string `validate:"required,oneof=dev prod"`
```

### Advanced Validation
Custom validation tags supported:
- `required_if_all_set=FieldA FieldB`
- `required_if_none_set=FieldA FieldB`
- `required_if_one_set=FieldA FieldB`
- `required_if_none_set_or_one_set=FieldA FieldB`
- `required_if_at_most_one_set=FieldA FieldB`
- `required_if_at_most_one_not_set=FieldA FieldB`

See `validator_test.go` for usage examples.

## Testing

Unit tests are provided in the loader-specific test files and `validator_test.go`. Run tests with:

```shell
go test ./...
```

### Running Benchmarks

To measure performance and memory usage, run benchmarks with:

```shell
go test -bench . -benchmem
```

This will report execution speed and memory allocations for all loaders and handlers, helping you track efficiency and identify areas for optimisation.

## License

MIT
