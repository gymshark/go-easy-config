package generic

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type YAMLLoader[T any] struct {
	Source interface{}
}

func (y *YAMLLoader[T]) Load(c *T) error {
	var data []byte
	var err error

	switch src := y.Source.(type) {
	case string:
		data, err = os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("error reading yaml file: %w", err)
		}
	case []byte:
		data = src
	default:
		return fmt.Errorf("unsupported source type: %T", src)
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("error unmarshalling yaml: %w", err)
	}
	return nil
}
