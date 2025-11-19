<div align="center">
  <img class="logo" src="logo.png" alt="go-easy-config logo" width="50%" />
</div>

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
    - [InterpolatingChainLoader (Variable Interpolation Support)](#interpolatingchainloader-variable-interpolation-support)
    - [Providing Your Own Loader](#providing-your-own-loader)
- [Variable Interpolation](#variable-interpolation)
  - [Syntax Reference](#syntax-reference)
  - [Common Use Cases](#common-use-cases)
  - [Troubleshooting](#troubleshooting)
- [Error Handling](#error-handling)
  - [Error Types Overview](#error-types-overview)
  - [Error Inspection Patterns](#error-inspection-patterns)
  - [Common Error Scenarios](#common-error-scenarios)
  - [Best Practices](#best-practices)
- [Validation](#validation)
  - [Advanced Validation](#advanced-validation)
- [Testing](#testing)
- [License](#license)

## Features
- Load configuration from environment variables
- Parse command-line flags
- Fetch secrets from AWS Secrets Manager (optional)
- Load configuration from INI, JSON, and YAML files or byte arrays
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
	"github.com/gymshark/go-easy-config" // Core package including chain loaders and interfaces
	
	// (OPTIONAL) Import sub packagaes for custom loaders and validators
	awsloaders "github.com/gymshark/go-easy-config/loader/aws" // Provides AWS specific loaders
	genericloaders "github.com/gymshark/go-easy-config/loader/generic" // Provides generic loaders (environment, command-line, ini, json, yaml)
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
	// You can also use the LoadAndValidate method:
	if err := handler.LoadAndValidate(&cfg); err != nil {
		panic(err)
	}
	// Use cfg...
}
```

### AWS Secrets Manager Integration

To fetch secrets, add fields with the `secretfetch` tag and configure AWS credentials:

```go
import (
	awsloaders "github.com/gymshark/go-easy-config/loader/aws"
)

type SecretsConfig struct {
	DBPassword string `secret:"aws=prod/db/password"`
}

// You'll also need to implement the handler with the AWS Secrets Manager loader:
handler := config.NewConfigHandler[SecretsConfig](
	config.WithLoaders(awsloaders.SecretsManagerLoader[SecretsConfig]{}, config.DefaultConfigLoaders()...),
)
```

### Customising Loaders and Validators

You can provide custom loaders or validators:

```go
handler := config.NewConfigHandler[AppConfig](
	config.WithValidator[AppConfig](customValidator),
	config.WithLoaders[AppConfig](customLoader),
)
```

### Types of Configuration Sources

#### Environment Variables (`env` tag)
Fields tagged with `env:"NAME"` are loaded from environment variables using [caarlos0/env](https://github.com/caarlos0/env).

#### Command-Line Arguments (`clap` tag)
Fields tagged with `clap:"name"` are loaded from command-line flags using [go-clap](https://github.com/fred1268/go-clap).

#### AWS Secrets Manager (`secret` tag)
Fields tagged with `secret:"aws=path/to/secret"` are loaded from AWS Secrets Manager using [secretfetch](https://github.com/crazywolf132/secretfetch).

#### INI Files or Byte Arrays (`ini` tag)
Fields can be loaded from INI files or byte arrays using [go-ini/ini](https://github.com/go-ini/ini).

#### JSON Files or Byte Arrays (`json` tag)
Fields can be loaded from JSON files or byte arrays using the Go standard library [encoding/json](https://pkg.go.dev/encoding/json).

#### YAML Files or Byte Arrays (`yaml` tag)
Fields can be loaded from YAML files or byte arrays using [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3).

### Loader Order and Customisation

By default, the configuration is loaded in the following order:
1. Environment variables
2. Command-line arguments

Each loader processes the configuration struct in sequence. If a variable is set by an earlier loader, it may be overridden by a later loader if the same field is present in multiple sources. For example, a value set via an environment variable will be replaced if a command-line argument for the same field is provided, and both may be replaced if a secret is fetched from AWS Secrets Manager for that field.

You can customise the loader order and override loaders using `WithLoaders`.

#### Custom Loader Order Example

To specify your own loader order:

```go
package main

import (
	"os"
	
	"github.com/gymshark/go-easy-config"
	"github.com/gymshark/go-easy-config/loader/generic"
	"github.com/gymshark/go-easy-config/loader/aws"
)

// Pass in individual loaders
// These are automatically wrapped in InterpolatingChainLoader
handler := config.NewConfigHandler[AppConfig](
	config.WithLoaders(
        &generic.CommandLineLoader[AppConfig]{Args: os.Args[1:]},
        &generic.EnvironmentLoader[AppConfig]{},
        &aws.SecretsManagerLoader[AppConfig]{},
    ),
)

```

> **Note:** `WithLoaders()` automatically wraps your loaders in an `InterpolatingChainLoader`, which provides variable interpolation support while maintaining the loader order you specify. This means variable interpolation works automatically with any custom loader configuration.


#### InterpolatingChainLoader (Variable Interpolation Support)

The `InterpolatingChainLoader` is used **automatically by default** when you call `NewConfigHandler()` or `WithLoaders()`. It provides variable interpolation support while maintaining full backward compatibility.

**You typically don't need to use it explicitly**, but you can if you need advanced features:

```go
// Automatic usage (recommended for 99% of cases)
handler := config.NewConfigHandler[AppConfig](
    config.WithLoaders(
        &generic.EnvironmentLoader[AppConfig]{},
        &generic.CommandLineLoader[AppConfig]{Args: os.Args[1:]},
    ),
)
// Variable interpolation works automatically!

// Explicit usage (only needed for advanced scenarios)
interpolatingLoader := &config.InterpolatingChainLoader[AppConfig]{
    Loaders: []config.Loader[AppConfig]{
        &generic.EnvironmentLoader[AppConfig]{},
        &generic.CommandLineLoader[AppConfig]{Args: os.Args[1:]},
    },
    ShortCircuit: true, // Enable short-circuit behavior
}

handler := config.NewConfigHandler[AppConfig](
    config.WithLoaders(interpolatingLoader),
)

// Access interpolation context for debugging
var cfg AppConfig
handler.Load(&cfg)
context := interpolatingLoader.GetInterpolationContext()
// Returns map[string]string of resolved variables
```

**When to use explicit `InterpolatingChainLoader`:**
- Enable `ShortCircuit: true` to stop loading when all fields are populated (performance optimization)
- Access `GetInterpolationContext()` to inspect resolved variables for debugging or logging

**Features:**
- Automatic dependency analysis and cycle detection
- Staged loading (fields loaded in dependency order)
- Zero overhead when no interpolation is used (fast path detection)
- Works with all loader types (env, CLI, AWS, file loaders)

See the [Variable Interpolation](#variable-interpolation) section for usage examples.

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
loaders := []loader.Loader[AppConfig]{
	&FileLoader[AppConfig]{Path: "config.yaml"},
	&generic.EnvironmentLoader[AppConfig]{},
}
handler := config.NewConfigHandler[AppConfig](
	config.WithLoaders(loaders...),
)
```

## Variable Interpolation

Variable interpolation allows you to reference field values within other field annotations using `${VARIABLE_NAME}` syntax. This enables dynamic configuration paths based on runtime context, such as environment-specific AWS Secrets Manager paths or configuration file names.

**Variable interpolation works automatically** with `NewConfigHandler()` and requires no special configuration. See [InterpolatingChainLoader](#interpolatingchainloader-variable-interpolation-support) for advanced usage options.

### Syntax Reference

#### Declaring Variables with `availableAs`

To make a field's value available for interpolation, use the `config:"availableAs=NAME"` tag:

```go
type Config struct {
    Environment string `env:"ENV" config:"availableAs=ENV" validate:"required"`
}
```

**Rules:**
- Variable names must contain only alphanumeric characters, underscores, and hyphens
- Each `availableAs` name must be unique across the struct
- Fields with `availableAs` must be exported (start with uppercase letter)
- Supported types: `string`, `int` (all variants), `uint` (all variants), `float32`, `float64`, `bool`

#### Referencing Variables with `${VAR}`

Reference declared variables in any struct tag using `${VARIABLE_NAME}` syntax:

```go
type Config struct {
    Environment string `env:"ENV" config:"availableAs=ENV"`
    DBPassword  string `secret:"aws=/myapp/${ENV}/db/password"`
}
```

**Features:**
- Multiple variables in a single tag: `${VAR1}/path/${VAR2}`
- Adjacent variables: `${VAR1}${VAR2}` (concatenated)
- Works with all loader types (env, CLI, AWS, file loaders)

### Common Use Cases

#### Environment-Based AWS Secrets Manager Paths

```go
type Config struct {
    Environment string `env:"ENV" config:"availableAs=ENV" validate:"required,oneof=dev staging prod"`
    DBPassword  string `secret:"aws=/myapp/${ENV}/db/password" validate:"required"`
    APIKey      string `secret:"aws=/myapp/${ENV}/api-key" validate:"required"`
}
```

When `ENV=prod`, the secrets are fetched from:
- `/myapp/prod/db/password`
- `/myapp/prod/api-key`

#### Multi-Variable Interpolation

```go
type Config struct {
    Env    string `env:"ENV" config:"availableAs=ENV"`
    Region string `env:"REGION" config:"availableAs=REGION"`
    Secret string `secret:"aws=/myapp/${ENV}/${REGION}/secret"`
}
```

With `ENV=prod` and `REGION=us-east-1`, the secret path becomes:
- `/myapp/prod/us-east-1/secret`

#### Numeric Types as Variables

```go
type Config struct {
    Port       int    `env:"PORT" config:"availableAs=PORT" validate:"required"`
    ConfigFile string `yaml:"config-${PORT}.yaml"`
}
```

With `PORT=8080`, the YAML file loaded is:
- `config-8080.yaml`

#### Dependency Chains

```go
type Config struct {
    // Stage 0: No dependencies - loaded first
    Env string `env:"ENV" config:"availableAs=ENV"`
    
    // Stage 1: Depends on ENV - loaded second
    Region string `env:"REGION_${ENV}" config:"availableAs=REGION"`
    
    // Stage 2: Depends on both ENV and REGION - loaded last
    Secret string `secret:"aws=/${ENV}/${REGION}/secret"`
}
```

The system automatically determines the correct loading order based on dependencies.

### Troubleshooting

#### Undefined Variable Error

**Symptom:**
```
undefined variable '${VAR}' referenced in field 'FieldName'
```

**Cause:** Variable referenced but no field has `config:"availableAs=VAR"`

**Solution:** Add `config:"availableAs=VAR"` to the field providing the value

#### Cyclic Dependency Error

**Symptom:**
```
cyclic dependency detected: FieldA -> FieldB -> FieldA
```

**Cause:** Fields depend on each other in a circular manner

**Solution:** Restructure configuration to break the cycle. Fields cannot depend on each other circularly.

#### Duplicate availableAs Error

**Symptom:**
```
duplicate availableAs='VAR' declared in fields: Field1, Field2
```

**Cause:** Multiple fields declare the same variable name

**Solution:** Use unique names for each `availableAs` declaration

#### Unsupported Type Error

**Symptom:**
```
unsupported type for interpolation: struct
```

**Cause:** Field with `availableAs` is a complex type (struct, slice, map, pointer)

**Solution:** Only use simple types (string, int, bool, float) for fields with `availableAs`

#### Non-Exported Field Error

**Symptom:**
```
field with availableAs must be exported
```

**Cause:** Field with `availableAs` starts with lowercase letter

**Solution:** Rename field to start with uppercase letter (e.g., `env` → `Env`)

## Error Handling

The library provides structured error types that enable programmatic error inspection and handling. All custom error types support Go's standard error wrapping and inspection using `errors.As` and `errors.Is`.

### Error Types Overview

| Error Type | Returned By | Description |
|------------|-------------|-------------|
| `LoaderError` | All loaders | Configuration loading failures (file read, parse, AWS errors) |
| `ValidationError` | `Handler.Validate()` | Validation rule violations |
| `TagParseError` | Tag parsing | Malformed struct tags |
| `InterpolationError` | Interpolation engine | Variable interpolation failures |
| `CyclicDependencyError` | Dependency analysis | Circular field dependencies |
| `UndefinedVariableError` | Dependency analysis | References to non-existent variables |
| `DuplicateAvailableAsError` | Dependency analysis | Duplicate variable declarations |
| `DependencyGraphError` | Dependency graph | General dependency graph failures |

### Error Inspection Patterns

#### Checking for Specific Error Types

Use `errors.As` to check for and extract specific error types:

```go
handler := config.NewConfigHandler[AppConfig]()
var cfg AppConfig

if err := handler.LoadAndValidate(&cfg); err != nil {
    // Check for loader errors
    var loaderErr *config.LoaderError
    if errors.As(err, &loaderErr) {
        fmt.Printf("Loader '%s' failed during %s\n", loaderErr.LoaderType, loaderErr.Operation)
        if loaderErr.Source != "" {
            fmt.Printf("Source: %s\n", loaderErr.Source)
        }
        // Access underlying error
        fmt.Printf("Details: %v\n", loaderErr.Err)
    }

    // Check for validation errors
    var validationErr *config.ValidationError
    if errors.As(err, &validationErr) {
        fmt.Printf("Field '%s' failed validation rule '%s'\n",
            validationErr.FieldName, validationErr.Rule)
    }

    // Check for interpolation errors
    var interpErr *config.InterpolationError
    if errors.As(err, &interpErr) {
        fmt.Printf("Interpolation failed in field '%s': %s\n",
            interpErr.FieldName, interpErr.Message)
    }
}
```

#### Handling Loader-Specific Errors

Different loaders return `LoaderError` with different contexts:

```go
if err := handler.Load(&cfg); err != nil {
    var loaderErr *config.LoaderError
    if errors.As(err, &loaderErr) {
        switch loaderErr.LoaderType {
        case "JSONLoader", "YAMLLoader", "INILoader":
            // File loading errors
            fmt.Printf("Failed to load file: %s\n", loaderErr.Source)
            
            // Check for specific file errors
            var pathErr *os.PathError
            if errors.As(loaderErr.Err, &pathErr) {
                fmt.Printf("File not found: %s\n", pathErr.Path)
            }
            
            var syntaxErr *json.SyntaxError
            if errors.As(loaderErr.Err, &syntaxErr) {
                fmt.Printf("JSON syntax error at offset %d\n", syntaxErr.Offset)
            }

        case "SecretsManagerLoader", "SSMParameterStoreLoader":
            // AWS errors
            fmt.Printf("AWS operation failed: %v\n", loaderErr.Err)

        case "EnvironmentLoader":
            // Environment variable parsing errors
            fmt.Printf("Environment variable parsing failed: %v\n", loaderErr.Err)

        case "CommandLineLoader":
            // Command-line argument parsing errors
            fmt.Printf("Command-line parsing failed: %v\n", loaderErr.Err)
        }
    }
}
```

#### Handling Variable Interpolation Errors

Variable interpolation can fail in several ways:

```go
if err := handler.Load(&cfg); err != nil {
    // Undefined variable
    var undefErr *config.UndefinedVariableError
    if errors.As(err, &undefErr) {
        fmt.Printf("Variable '${%s}' not found in field '%s'\n",
            undefErr.VariableName, undefErr.FieldName)
        fmt.Printf("Fix: Add config:\"availableAs=%s\" to the field providing this value\n",
            undefErr.VariableName)
    }

    // Circular dependency
    var cycleErr *config.CyclicDependencyError
    if errors.As(err, &cycleErr) {
        fmt.Printf("Circular dependency: %s\n", strings.Join(cycleErr.Cycle, " -> "))
        fmt.Printf("Fix: Break the dependency chain by restructuring your configuration\n")
    }

    // Duplicate variable declaration
    var dupErr *config.DuplicateAvailableAsError
    if errors.As(err, &dupErr) {
        fmt.Printf("Duplicate variable '%s' in fields: %s\n",
            dupErr.VariableName, strings.Join(dupErr.Fields, ", "))
        fmt.Printf("Fix: Use unique names for each availableAs declaration\n")
    }

    // Tag parsing error
    var tagErr *config.TagParseError
    if errors.As(err, &tagErr) {
        fmt.Printf("Tag parse error in field '%s' (tag: %s): %s\n",
            tagErr.FieldName, tagErr.TagKey, tagErr.Issue)
    }
}
```

#### Distinguishing Load vs Validation Failures

The `LoadAndValidate` method can fail at different stages:

```go
if err := handler.LoadAndValidate(&cfg); err != nil {
    var loaderErr *config.LoaderError
    var validationErr *config.ValidationError

    if errors.As(err, &loaderErr) {
        // Failed during loading
        fmt.Printf("Configuration loading failed: %v\n", err)
        // Handle loader-specific errors
    } else if errors.As(err, &validationErr) {
        // Failed during validation (loading succeeded)
        fmt.Printf("Configuration validation failed: %v\n", err)
        // Handle validation errors
    }
}
```

#### Accessing Wrapped Errors

All error types that wrap underlying errors support `errors.Unwrap`:

```go
if err := handler.Load(&cfg); err != nil {
    var loaderErr *config.LoaderError
    if errors.As(err, &loaderErr) {
        // Access the underlying error directly
        underlyingErr := errors.Unwrap(loaderErr)
        // Or use the Err field
        underlyingErr = loaderErr.Err

        // Check for specific underlying error types
        var pathErr *os.PathError
        if errors.As(underlyingErr, &pathErr) {
            fmt.Printf("File operation failed: %s\n", pathErr.Op)
        }
    }
}
```

### Common Error Scenarios

#### File Not Found

```go
// Scenario: JSON file doesn't exist
// Error: JSONLoader error during read file (source: config.json): open config.json: no such file or directory

var loaderErr *config.LoaderError
if errors.As(err, &loaderErr) && loaderErr.LoaderType == "JSONLoader" {
    var pathErr *os.PathError
    if errors.As(loaderErr.Err, &pathErr) {
        fmt.Printf("Configuration file not found: %s\n", pathErr.Path)
    }
}
```

#### Invalid JSON Syntax

```go
// Scenario: JSON file has syntax errors
// Error: JSONLoader error during unmarshal JSON (source: config.json): invalid character '}' after object key

var loaderErr *config.LoaderError
if errors.As(err, &loaderErr) && loaderErr.Operation == "unmarshal JSON" {
    var syntaxErr *json.SyntaxError
    if errors.As(loaderErr.Err, &syntaxErr) {
        fmt.Printf("JSON syntax error at byte offset %d\n", syntaxErr.Offset)
    }
}
```

#### Validation Failure

```go
// Scenario: Port value is out of range
// Error: validation failed for field 'Port': rule 'min=1' failed

var validationErr *config.ValidationError
if errors.As(err, &validationErr) {
    fmt.Printf("Field '%s' failed validation\n", validationErr.FieldName)
    fmt.Printf("Rule: %s\n", validationErr.Rule)
    if validationErr.Value != "" {
        fmt.Printf("Invalid value: %s\n", validationErr.Value)
    }
}
```

#### Undefined Variable

```go
// Scenario: Field references ${ENV} but no field declares availableAs=ENV
// Error: undefined variable '${ENV}' referenced in field 'DatabaseURL'

var undefErr *config.UndefinedVariableError
if errors.As(err, &undefErr) {
    fmt.Printf("Add this to your config struct:\n")
    fmt.Printf("  Environment string `env:\"ENV\" config:\"availableAs=%s\"`\n",
        undefErr.VariableName)
}
```

#### Circular Dependency

```go
// Scenario: FieldA depends on FieldB, FieldB depends on FieldA
// Error: cyclic dependency detected: FieldA -> FieldB -> FieldA

var cycleErr *config.CyclicDependencyError
if errors.As(err, &cycleErr) {
    fmt.Printf("Circular dependency detected:\n")
    for i, field := range cycleErr.Cycle {
        if i > 0 {
            fmt.Printf(" -> ")
        }
        fmt.Printf("%s", field)
    }
    fmt.Println("\nRestructure your configuration to break this cycle")
}
```

### Best Practices

1. **Always check errors**: Never ignore errors from `Load()`, `Validate()`, or `LoadAndValidate()`
2. **Use errors.As for inspection**: Prefer `errors.As` over type assertions for error checking
3. **Check specific types first**: Check for more specific error types before generic ones
4. **Access wrapped errors**: Use `errors.Unwrap()` or the `Err` field to access underlying errors
5. **Provide context in logs**: Include error type, field names, and operation details in log messages
6. **Handle gracefully**: Provide helpful error messages to users based on error type

## Validation

Configuration structs can be validated using [go-playground/validator](https://github.com/go-playground/validator). Use the `validate` struct tag to specify rules:

```go
Port int `validate:"required,min=1,max=65535"`
Env  string `validate:"required,oneof=dev prod"`
```

### Advanced Validation

Custom validation tags supported:

- `required_if_all_set=FieldA FieldB`
  Ensures the field is required if all listed fields are set (non-zero).
- `required_if_none_set=FieldA FieldB`
  Ensures the field is required if none of the listed fields are set (all zero).
- `required_if_one_set=FieldA FieldB`
  Ensures the field is required if exactly one of the listed fields is set (non-zero).
- `required_if_none_set_or_one_set=FieldA FieldB`
  Ensures the field is required if either none or exactly one of the listed fields is set.
- `required_if_at_most_one_set=FieldA FieldB`
  Ensures the field is required if at most one of the listed fields is set (zero or one).
- `required_if_at_most_one_not_set=FieldA FieldB`
  Ensures the field is required if at most one of the listed fields is not set (zero or one unset).

These tags allow for conditional validation logic based on the state of other fields in the struct. For example, you can require a field only if certain other fields are present or absent, supporting complex configuration requirements.

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

## License

MIT
