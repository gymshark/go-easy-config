package generic

import (
	"encoding/json"
	"fmt"
	"os"
)

// JSONLoader loads configuration from JSON files or byte arrays.
type JSONLoader[T any] struct {
	Source interface{} // Either a file path (string) or raw JSON data ([]byte)
}

// Load populates configuration from JSON source.
func (j *JSONLoader[T]) Load(c *T) error {
	var data []byte
	var err error

	switch src := j.Source.(type) {
	case string:
		// Assume it's a file path
		data, err = os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("error reading json file: %w", err)
		}
	case []byte:
		data = src
	default:
		return fmt.Errorf("unsupported source type: %T", src)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("error unmarshalling json: %w", err)
	}
	return nil
}
