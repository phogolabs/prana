package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/svett/gom"
	"github.com/urfave/cli"
)

type SQLCommand struct {
	generator *gom.CmdGenerator
}

func (m *SQLCommand) CreateCommand() cli.Command {
	return cli.Command{
		Name:         "command",
		Usage:        "A group of commands for generating, running, and removing SQL commands",
		Description:  "A group of commands for generating, running, and removing SQL commands",
		BashComplete: cli.DefaultAppComplete,
		Before:       m.BeforeEach,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "create",
				Usage:       "Create a new command for given container filename",
				Description: "Create a new command for given container filename",
				ArgsUsage:   "[container] [name]",
				Action:      m.Create,
			},
		},
	}
}

func (m *SQLCommand) BeforeEach(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	dir = filepath.Join(dir, "/database/command")

	m.generator = &gom.CmdGenerator{
		Dir: dir,
	}

	return nil
}

func (m *SQLCommand) Create(ctx *cli.Context) error {
	args := ctx.Args()

	if len(args) != 2 {
		return cli.NewExitError("Create command expects two arguments", ErrCodeCommand)
	}

	container := strings.Replace(args[0], " ", "_", -1)
	name := strings.Replace(args[1], " ", "-", -1)
	path, err := m.generator.Create(container, name)
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeCommand)
	}

	log.Infof("Command '%s' has been created at '%s' successfully", name, path)
	return nil
}
