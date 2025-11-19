package generic

import (
	"fmt"
	"os"

	"github.com/gymshark/go-easy-config/loader"
	"gopkg.in/yaml.v3"
)

// YAMLLoader loads configuration from YAML files or byte arrays.
type YAMLLoader[T any] struct {
	Source interface{} // Either a file path (string) or raw YAML data ([]byte)
}

// Load populates configuration from YAML source.
func (y *YAMLLoader[T]) Load(c *T) error {
	var data []byte
	var err error
	var source string

	switch src := y.Source.(type) {
	case string:
		source = src
		data, err = os.ReadFile(src)
		if err != nil {
			return &loader.LoaderError{
				LoaderType: "YAMLLoader",
				Operation:  "read file",
				Source:     source,
				Err:        err,
			}
		}
	case []byte:
		data = src
		source = "<bytes>"
	default:
		return &loader.LoaderError{
			LoaderType: "YAMLLoader",
			Operation:  "validate source type",
			Source:     fmt.Sprintf("%T", src),
			Err:        fmt.Errorf("unsupported source type"),
		}
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		return &loader.LoaderError{
			LoaderType: "YAMLLoader",
			Operation:  "unmarshal YAML",
			Source:     source,
			Err:        err,
		}
	}
	return nil
}
