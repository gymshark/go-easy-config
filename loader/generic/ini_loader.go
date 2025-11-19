package generic

import (
	"fmt"

	"github.com/gymshark/go-easy-config/loader"
	"gopkg.in/ini.v1"
)

// IniLoader loads configuration from INI files or byte arrays.
type IniLoader[T any] struct {
	Source      interface{}     // Either a file path (string) or raw INI data ([]byte)
	LoadOptions ini.LoadOptions // Options for INI parsing
	INI         *ini.File       // Parsed INI file data structure (populated after Load)
}

// Load populates configuration from INI source using struct tags.
func (i *IniLoader[T]) Load(c *T) error {
	var source string
	switch src := i.Source.(type) {
	case string:
		source = src
	case []byte:
		source = "<bytes>"
	default:
		source = fmt.Sprintf("%T", src)
	}

	data, err := ini.LoadSources(i.LoadOptions, i.Source)
	if err != nil {
		return &loader.LoaderError{
			LoaderType: "INILoader",
			Operation:  "load INI file",
			Source:     source,
			Err:        err,
		}
	}

	i.INI = data

	if err := data.MapTo(c); err != nil {
		return &loader.LoaderError{
			LoaderType: "INILoader",
			Operation:  "map INI to struct",
			Source:     source,
			Err:        err,
		}
	}

	return nil
}
