# go-easy-config Agent Instructions

A lightweight configuration loader and validator library for Go applications. Supports loading configuration from environment variables, command-line flags, AWS Secrets Manager, and file formats (INI, JSON, YAML) with built-in validation.

## Project Context

### Type
Library project - no main application to run. The primary deliverable is the Go package that other applications import for configuration management.

### Repository Structure
```
├── config.go                         # Main configuration handler with generics
├── config_test.go                    # Core functionality tests
├── interpolating_chain_loader.go     # Chain loader with interpolation and short-circuit support
├── interpolation.go                  # Variable interpolation engine
├── interpolation_test.go             # Interpolation tests
├── interpolation_errors.go           # Custom error types for interpolation
├── tag_parser.go                     # Tag parsing utilities
├── tag_parser_test.go                # Tag parser tests
├── dependency_graph.go               # Dependency graph and topological sort
├── dependency_graph_test.go          # Dependency graph tests
├── validator.go                      # Custom validation rules
├── loader/
│   ├── generic/                      # Standard loaders (env, CLI, INI, JSON, YAML)
│   └── aws/                          # AWS integration loaders (Secrets Manager, SSM)
├── utils/                            # Utility functions
└── Makefile                          # Build automation
```

### Key Dependencies
- `github.com/caarlos0/env/v11` - Environment variable parsing
- `github.com/fred1268/go-clap` - Command-line argument parsing
- `github.com/go-playground/validator/v10` - Struct validation
- `github.com/crazywolf132/secretfetch` - AWS Secrets Manager integration
- `gopkg.in/ini.v1`, `gopkg.in/yaml.v3` - File format support

### Configuration Load Order (Default)
1. Environment variables (highest precedence)
2. Command-line arguments (overrides environment)
3. AWS Secrets Manager (if configured)
4. File sources (lowest precedence)

Later loaders override earlier ones for the same field.

### Variable Interpolation
The library supports dynamic variable interpolation in struct tags using `${VAR}` syntax. Fields can be declared as variables using `config:"availableAs=NAME"` and referenced in other field tags. The system automatically:
- Analyzes dependencies between fields
- Detects circular dependencies
- Loads fields in dependency-ordered stages
- Supports string, int, uint, float, and bool types as variables
- Works with all loader types (env, CLI, AWS, file loaders)

## Development Workflow

### Initial Setup
```bash
cd /home/runner/work/go-easy-config/go-easy-config
make setup  # Downloads dependencies, ~7 seconds. NEVER CANCEL. Set timeout to 30+ seconds
# Alternative: go mod tidy
```

### Building and Testing
```bash
# Run tests with race detection (recommended)
make test  # ~30 seconds. NEVER CANCEL. Set timeout to 60+ seconds

# Run tests quickly without race detection
go test ./...  # ~4 seconds. Set timeout to 15+ seconds

# Run specific package tests
go test ./loader/generic -v  # Test generic loaders only
go test ./loader/aws -v      # Test AWS loaders only

# Run benchmarks
make test-bench  # ~27 seconds. NEVER CANCEL. Set timeout to 60+ seconds
# Alternative: go test -bench . -benchmem
```

### Code Quality
```bash
# Format code
make fmt  # ~0.4 seconds
# Alternative: gofmt -s -w -l .

# Static analysis
go vet ./...  # ~1 second
```

### Pre-commit Checklist
Always run before completing work:
1. `make fmt` - format code
2. `go vet ./...` - static analysis
3. `make test` - full test suite with race detection
4. Create and run a validation scenario for your changes

## Validation and Testing

### Manual Testing Pattern
ALWAYS test changes by creating a validation scenario:

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

Run with: `go run /tmp/test_scenario.go`

Expected output: Configuration values printed successfully

## Common Development Patterns

### Configuration Struct Pattern
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

### Variable Interpolation Pattern
```go
type AppConfig struct {
    // Declare variables with availableAs
    Environment string `env:"ENV" config:"availableAs=ENV" validate:"required,oneof=dev staging prod"`
    Region      string `env:"REGION" config:"availableAs=REGION" validate:"required"`
    
    // Reference variables in other tags using ${VAR} syntax
    DBPassword  string `secret:"aws=/myapp/${ENV}/db/password" validate:"required"`
    APIKey      string `secret:"aws=/myapp/${ENV}/${REGION}/api-key" validate:"required"`
    
    // Numeric types work too
    Port        int    `env:"PORT" config:"availableAs=PORT" validate:"required"`
    ConfigFile  string `yaml:"config-${PORT}.yaml"`
}
```

**Key Points:**
- Use `config:"availableAs=NAME"` to declare a field as a variable
- Reference variables with `${NAME}` in any struct tag
- System automatically determines loading order based on dependencies
- Supports string, int, uint, float, and bool types
- Works with all loader types (env, CLI, AWS, file loaders)

### Creating New Loaders
1. Implement the `Loader[T]` interface in appropriate package (`loader/generic/` or `loader/aws/`)
2. Add corresponding test file following existing patterns (`*_test.go`)
3. Update default loader chain in `config.go` if needed

### Adding Validation Rules
1. Register new validation rules in `validator.go` `NewValidator()` function
2. Add test cases in `validator_test.go`
3. Follow pattern: `validate.RegisterValidation("rule_name", func(fl validator.FieldLevel) bool { ... })`

### Testing Changes
1. Always add unit tests following existing patterns in `*_test.go` files
2. Test with multiple loaders: `NewConfigHandler(WithLoaders(loader1, loader2))`
3. Test validation scenarios with both valid and invalid data
4. Benchmark performance-critical changes with `go test -bench .`

### Testing Variable Interpolation
When testing interpolation features:
1. Test dependency analysis with various struct configurations
2. Test cycle detection with circular dependencies
3. Test undefined variable detection
4. Test duplicate availableAs detection
5. Test type conversion for all supported types (string, int, uint, float, bool)
6. Test multi-variable interpolation in single tags
7. Test staged loading with mock loaders
8. Test integration with all loader types

## Navigation Guide

- **Main API**: Start with `config.go` - contains `NewConfigHandler[T]()` and core methods
- **Loader implementations**: Check `loader/generic/` for standard sources, `loader/aws/` for AWS sources
- **Validation logic**: Custom validation rules are in `validator.go`
- **Test patterns**: Look at `*_test.go` files for usage examples and patterns
- **Chain loading**: `interpolating_chain_loader.go` shows how multiple sources are combined with interpolation support
- **Variable interpolation**: 
  - `interpolation.go` - Core interpolation engine with dependency analysis
  - `interpolating_chain_loader.go` - Staged loading with interpolation support
  - `tag_parser.go` - Variable reference extraction and string interpolation
  - `dependency_graph.go` - Dependency graph construction and topological sort
  - `interpolation_errors.go` - Custom error types for interpolation failures

## Troubleshooting

### Common Issues
- **Import errors**: Run `make setup` to ensure all dependencies are downloaded
- **Test failures**: Check environment variables are not conflicting with test values
- **Validation errors**: Verify struct tags match the validation rules in `validator.go`
- **AWS loader issues**: Ensure AWS credentials are configured if testing AWS functionality
- **Interpolation errors**:
  - **Undefined variable**: Add `config:"availableAs=VAR"` to the field providing the value
  - **Cyclic dependency**: Restructure configuration to break circular references
  - **Duplicate availableAs**: Use unique names for each variable declaration
  - **Unsupported type**: Only use simple types (string, int, bool, float) for availableAs fields
  - **Non-exported field**: Rename field to start with uppercase letter

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

## Important Notes

- Always reference these instructions first and fallback to search or bash commands only when encountering unexpected information
- This is a library project - focus on API design, testing, and documentation
- Never cancel long-running commands like `make setup` or `make test` - they need time to complete
- Always create validation scenarios to test changes to core functionality
