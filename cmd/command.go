package cmd

import (
	"github.com/bowouhuy/bookstore-api/config"
	"github.com/urfave/cli/v2"
)

// Command is a struct
type Command struct {
	*config.Config

	CLI []*cli.Command
}

// NewCommand is a constructor
func NewCommand(options ...Option) *Command {
	cmd := &Command{}

	for _, op := range options {
		op(cmd)
	}

	return cmd
}

// registerCLI is a function
func (cmd *Command) registerCLI(cmdCLI *cli.Command) {
	cmd.CLI = append(cmd.CLI, cmdCLI)
}
