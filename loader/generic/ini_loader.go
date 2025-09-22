package generic

import (
	"fmt"

	"gopkg.in/ini.v1"
)

// IniLoader loads configuration from INI files or byte arrays.
type IniLoader[T any] struct {
	Source      interface{}       // Either a file path (string) or raw INI data ([]byte)
	LoadOptions ini.LoadOptions   // Options for INI parsing
	INI         *ini.File         // Parsed INI file data structure (populated after Load)
}

// Load populates configuration from INI source using struct tags.
func (i *IniLoader[T]) Load(c *T) error {
	data, err := ini.LoadSources(i.LoadOptions, i.Source)
	if err != nil {
		return fmt.Errorf("error loading ini file: %w", err)
	}

	i.INI = data

	return data.MapTo(c)
}
