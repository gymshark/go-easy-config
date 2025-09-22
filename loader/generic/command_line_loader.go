package generic

import (
	"fmt"

	"github.com/fred1268/go-clap/clap"
)

// CommandLineLoader loads configuration from command-line arguments.
// It supports fields tagged with `clap:"flag-name"`.
type CommandLineLoader[T any] struct {
	Args []string // Command-line arguments to parse (typically os.Args[1:])
}

// Load populates configuration fields from command-line arguments.
func (cmd *CommandLineLoader[T]) Load(c *T) error {
	_, err := clap.Parse(cmd.Args, c)
	if err != nil {
		return fmt.Errorf("error parsing command line arguments: %w", err)
	}
	return nil
}
