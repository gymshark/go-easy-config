package loader

import "fmt"

// LoaderError represents errors that occur during configuration loading from any loader.
// It provides context about which loader failed, what operation was being performed,
// and optionally the source being loaded (e.g., file path, environment variable name).
//
// Fields:
//   - LoaderType: Type of loader that failed (e.g., "JSONLoader", "EnvironmentLoader", "SecretsManagerLoader")
//   - Operation: Operation being performed (e.g., "read file", "unmarshal", "parse", "fetch secrets")
//   - Source: Optional source identifier (e.g., file path, environment variable name, secret path)
//   - Err: Underlying error that caused the failure (e.g., os.PathError, json.SyntaxError, AWS SDK errors)
//
// Loaders that return LoaderError:
//   - EnvironmentLoader - When parsing environment variables fails
//   - CommandLineLoader - When parsing command-line arguments fails
//   - JSONLoader - When reading or unmarshaling JSON files fails
//   - YAMLLoader - When reading or unmarshaling YAML files fails
//   - INILoader - When reading or parsing INI files fails
//   - SecretsManagerLoader - When AWS Secrets Manager operations fail
//   - SSMParameterStoreLoader - When AWS SSM Parameter Store operations fail
//
// Example - Creating a LoaderError:
//
//	err := &LoaderError{
//	    LoaderType: "JSONLoader",
//	    Operation:  "read file",
//	    Source:     "config.json",
//	    Err:        osErr,
//	}
//
// Example - Inspecting loader errors:
//
//	handler := config.NewConfigHandler[AppConfig]()
//	var cfg AppConfig
//	if err := handler.Load(&cfg); err != nil {
//	    var loaderErr *LoaderError
//	    if errors.As(err, &loaderErr) {
//	        fmt.Printf("Loader '%s' failed\n", loaderErr.LoaderType)
//	        fmt.Printf("Operation: %s\n", loaderErr.Operation)
//	        if loaderErr.Source != "" {
//	            fmt.Printf("Source: %s\n", loaderErr.Source)
//	        }
//	        // Access underlying error for more details
//	        fmt.Printf("Underlying error: %v\n", loaderErr.Err)
//	    }
//	}
//
// Example - Handling specific loader types:
//
//	if err := handler.Load(&cfg); err != nil {
//	    var loaderErr *LoaderError
//	    if errors.As(err, &loaderErr) {
//	        switch loaderErr.LoaderType {
//	        case "JSONLoader", "YAMLLoader":
//	            fmt.Printf("File loading failed: %s\n", loaderErr.Source)
//	        case "SecretsManagerLoader":
//	            fmt.Printf("AWS Secrets Manager error: %v\n", loaderErr.Err)
//	        case "EnvironmentLoader":
//	            fmt.Printf("Environment variable parsing failed\n")
//	        }
//	    }
//	}
//
// Example - Accessing wrapped errors:
//
//	if err := handler.Load(&cfg); err != nil {
//	    var loaderErr *LoaderError
//	    if errors.As(err, &loaderErr) {
//	        // Check for specific underlying error types
//	        var pathErr *os.PathError
//	        if errors.As(loaderErr.Err, &pathErr) {
//	            fmt.Printf("File not found: %s\n", pathErr.Path)
//	        }
//
//	        var syntaxErr *json.SyntaxError
//	        if errors.As(loaderErr.Err, &syntaxErr) {
//	            fmt.Printf("JSON syntax error at offset %d\n", syntaxErr.Offset)
//	        }
//	    }
//	}
type LoaderError struct {
	LoaderType string // Type of loader (e.g., "JSONLoader", "EnvironmentLoader")
	Operation  string // Operation being performed (e.g., "read file", "unmarshal", "parse")
	Source     string // Optional source identifier (e.g., file path, env var name)
	Err        error  // Underlying error that caused the failure
}

// Error returns a formatted error message with loader context.
// If Source is provided, it's included in the message for additional context.
func (e *LoaderError) Error() string {
	if e.Source != "" {
		return fmt.Sprintf("%s error during %s (source: %s): %v",
			e.LoaderType, e.Operation, e.Source, e.Err)
	}
	return fmt.Sprintf("%s error during %s: %v",
		e.LoaderType, e.Operation, e.Err)
}

// Unwrap returns the underlying error, enabling error chain traversal
// with errors.Is and errors.As.
func (e *LoaderError) Unwrap() error {
	return e.Err
}
