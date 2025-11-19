package generic

import (
	"github.com/fred1268/go-clap/clap"
	"github.com/gymshark/go-easy-config/loader"
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
		return &loader.LoaderError{
			LoaderType: "CommandLineLoader",
			Operation:  "parse command line arguments",
			Err:        err,
		}
	}
	return nil
}
