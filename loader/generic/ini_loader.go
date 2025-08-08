package generic

import (
	"fmt"

	"gopkg.in/ini.v1"
)

type IniLoader[T any] struct {
	// Source can be the raw ini data (byte array) or a path to a file
	Source      interface{}
	LoadOptions ini.LoadOptions

	// INI is the exposed INI file data structure
	INI *ini.File
}

func (i *IniLoader[T]) Load(c *T) error {
	data, err := ini.LoadSources(i.LoadOptions, i.Source)
	if err != nil {
		return fmt.Errorf("error loading ini file: %w", err)
	}

	i.INI = data

	return data.MapTo(c)
}
