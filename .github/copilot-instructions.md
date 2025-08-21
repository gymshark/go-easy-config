# go-easy-config

go-easy-config is a lightweight configuration loader and validator library for Go applications. It supports loading configuration from environment variables, command-line flags, AWS Secrets Manager, and file formats (INI, JSON, YAML) with built-in validation.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Initial Setup
- Bootstrap the repository and install dependencies:
  - `cd /home/runner/work/go-easy-config/go-easy-config`
  - `make setup` -- downloads dependencies, takes ~7 seconds. NEVER CANCEL. Set timeout to 30+ seconds.
  - Alternative: `go mod tidy` -- equivalent dependency setup command

### Building and Testing  
- Run tests with race detection (recommended):
  - `make test` -- runs verbose tests with race detection, takes ~30 seconds. NEVER CANCEL. Set timeout to 60+ seconds.
- Run tests quickly without race detection:
  - `go test ./...` -- runs all tests, takes ~4 seconds. Set timeout to 15+ seconds.
- Run specific package tests:
  - `go test ./loader/generic -v` -- test generic loaders only
  - `go test ./loader/aws -v` -- test AWS loaders only
- Run benchmarks:
  - `make test-bench` -- runs performance benchmarks, takes ~27 seconds. NEVER CANCEL. Set timeout to 60+ seconds.
  - `go test -bench . -benchmem` -- equivalent benchmark command

### Code Quality and Formatting
- Format code:
  - `make fmt` -- formats all Go code using gofmt, takes ~0.4 seconds
  - `gofmt -s -w -l .` -- equivalent formatting command
- Static analysis:
  - `go vet ./...` -- checks for common Go issues, takes ~1 second

## Validation

### Manual Testing Workflow
ALWAYS test your changes by creating a validation scenario like this:

```go
// Create in /tmp/test_scenario.go
package main

import (
    "fmt"
    "os"
    "github.com/gymshark/go-easy-config"
)

type TestConfig struct {
    DatabaseURL string `env:"DATABASE_URL" validate:"required"`
    Port        int    `env:"PORT" clap:"--port" validate:"required,min=1,max=65535"`
    Debug       bool   `env:"DEBUG" clap:"--debug"`
}

func main() {
    os.Setenv("DATABASE_URL", "postgresql://localhost:5432/testdb")
    os.Setenv("PORT", "8080")
    os.Setenv("DEBUG", "true")

    handler := config.NewConfigHandler[TestConfig]()
    var cfg TestConfig

    if err := handler.LoadAndValidate(&cfg); err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Configuration loaded successfully!\n")
    fmt.Printf("Database URL: %s\n", cfg.DatabaseURL)
    fmt.Printf("Port: %d\n", cfg.Port)
    fmt.Printf("Debug: %t\n", cfg.Debug)
}
```

- Run with: `go run /tmp/test_scenario.go`
- Expected output: Configuration values printed successfully
- ALWAYS create and run a validation scenario after making changes to core functionality

### Pre-commit Validation
Always run these commands before completing your work:
1. `make fmt` -- format code
2. `go vet ./...` -- static analysis  
3. `make test` -- full test suite with race detection
4. Create and run a validation scenario for your changes

## Repository Structure

### Key Directories and Files
```
/home/runner/work/go-easy-config/go-easy-config/
├── config.go                    # Main configuration handler with generics
├── config_test.go               # Core functionality tests
├── chain_loader.go              # Sequential loader implementation
├── short_circuit_chain_loader.go # Optimized loader with early exit
├── validator.go                 # Custom validation rules
├── loader/
│   ├── generic/                 # Standard loaders
│   │   ├── environment_loader.go    # env tag support
│   │   ├── command_line_loader.go   # clap tag support  
│   │   ├── ini_loader.go           # INI file support
│   │   ├── json_loader.go          # JSON file support
│   │   └── yaml_loader.go          # YAML file support
│   └── aws/                     # AWS integration loaders
│       ├── secrets_manager_loader.go    # secretfetch tag support
│       └── ssm_parameter_store_loader.go # SSM Parameter Store
├── utils/                       # Utility functions
└── Makefile                     # Build automation
```

### Navigation Tips
- **Main API**: Start with `config.go` - contains `NewConfigHandler[T]()` and core methods
- **Loader implementations**: Check `loader/generic/` for standard sources, `loader/aws/` for AWS sources
- **Validation logic**: Custom validation rules are in `validator.go`
- **Test patterns**: Look at `*_test.go` files for usage examples and patterns
- **Chain loading**: `chain_loader.go` shows how multiple sources are combined

## Common Development Patterns

### Creating New Loaders
1. Implement the `Loader[T]` interface in appropriate package (`loader/generic/` or `loader/aws/`)
2. Add corresponding test file following existing patterns (`*_test.go`)
3. Update default loader chain in `config.go` if needed

### Adding Validation Rules
1. Register new validation rules in `validator.go` `NewValidator()` function
2. Add test cases in `validator_test.go`
3. Follow pattern: `validate.RegisterValidation("rule_name", func(fl validator.FieldLevel) bool { ... })`

### Configuration Struct Patterns
```go
type AppConfig struct {
    // Environment variable with validation
    DatabaseURL string `env:"DATABASE_URL" validate:"required,url"`
    
    // Command line flag with env fallback
    Port int `env:"PORT" clap:"--port" validate:"required,min=1,max=65535"`
    
    // AWS Secrets Manager integration
    APIKey string `secretfetch:"/prod/api/key" validate:"required"`
    
    // File-based configuration
    Config FileConfig `ini:"config.ini" json:"config.json" yaml:"config.yaml"`
}
```

### Testing Changes
1. Always add unit tests following existing patterns in `*_test.go` files
2. Test with multiple loaders: `NewConfigHandler(WithLoaders(loader1, loader2))`
3. Test validation scenarios with both valid and invalid data
4. Benchmark performance-critical changes with `go test -bench .`

## Troubleshooting

### Common Issues
- **Import errors**: Run `make setup` to ensure all dependencies are downloaded
- **Test failures**: Check environment variables are not conflicting with test values
- **Validation errors**: Verify struct tags match the validation rules in `validator.go`
- **AWS loader issues**: Ensure AWS credentials are configured if testing AWS functionality

### Environment Requirements
- **Go version**: 1.24+ (currently using 1.24.6)
- **No external linters required**: Use built-in `go vet` and `gofmt`
- **No CI/CD**: All validation must be done locally before committing

### Performance Expectations
- Initial setup: ~7 seconds (first time with downloads)
- Full test suite: ~30 seconds (with race detection)
- Quick tests: ~4 seconds (without race detection) 
- Benchmarks: ~27 seconds
- Code formatting: ~0.4 seconds

## Key Project Information

This is a **library project** - there is no main application to run. The primary deliverable is the Go package that other applications import and use for configuration management.

### Dependencies
- `github.com/caarlos0/env/v11` - Environment variable parsing
- `github.com/fred1268/go-clap` - Command-line argument parsing  
- `github.com/go-playground/validator/v10` - Struct validation
- `github.com/crazywolf132/secretfetch` - AWS Secrets Manager integration
- `gopkg.in/ini.v1`, `gopkg.in/yaml.v3` - File format support

### Load Order (Default)
1. Environment variables (highest precedence)
2. Command-line arguments (overrides environment)
3. AWS Secrets Manager (if configured)
4. File sources (lowest precedence)

Later loaders override earlier ones for the same field.