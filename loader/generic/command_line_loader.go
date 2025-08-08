package generic

import (
	"fmt"

	"github.com/fred1268/go-clap/clap"
)

type CommandLineLoader[T any] struct {
	Args []string
}

func (cmd *CommandLineLoader[T]) Load(c *T) error {
	_, err := clap.Parse(cmd.Args, c)
	if err != nil {
		return fmt.Errorf("error parsing command line arguments: %w", err)
	}
	return nil
}
