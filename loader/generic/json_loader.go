package generic

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gymshark/go-easy-config/loader"
)

// JSONLoader loads configuration from JSON files or byte arrays.
type JSONLoader[T any] struct {
	Source interface{} // Either a file path (string) or raw JSON data ([]byte)
}

// Load populates configuration from JSON source.
func (j *JSONLoader[T]) Load(c *T) error {
	var data []byte
	var err error
	var source string

	switch src := j.Source.(type) {
	case string:
		source = src
		data, err = os.ReadFile(src)
		if err != nil {
			return &loader.LoaderError{
				LoaderType: "JSONLoader",
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
			LoaderType: "JSONLoader",
			Operation:  "validate source type",
			Source:     fmt.Sprintf("%T", src),
			Err:        fmt.Errorf("unsupported source type"),
		}
	}

	if err := json.Unmarshal(data, c); err != nil {
		return &loader.LoaderError{
			LoaderType: "JSONLoader",
			Operation:  "unmarshal JSON",
			Source:     source,
			Err:        err,
		}
	}
	return nil
}
