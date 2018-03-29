package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/olekukonko/tablewriter"
	"github.com/svett/gom"
	"github.com/svett/gom/script"
	"github.com/urfave/cli"
)

// SQLCommand provides a subcommands to work with SQL scripts and their
// statements.
type SQLCommand struct {
	dir string
}

// CreateCommand creates a cli.Command that can be used by cli.App.
func (m *SQLCommand) CreateCommand() cli.Command {
	return cli.Command{
		Name:         "command",
		Usage:        "A group of commands for generating, running, and removing SQL commands",
		Description:  "A group of commands for generating, running, and removing SQL commands",
		BashComplete: cli.DefaultAppComplete,
		Before:       m.beforeEach,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "create",
				Usage:       "Create a new SQL command for given container filename",
				Description: "Create a new SQL command for given container filename",
				ArgsUsage:   "[name]",
				Action:      m.create,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "filename, n",
						Usage: "Name of the file that contains the command",
						Value: "",
					},
				},
			},
			cli.Command{
				Name:        "run",
				Usage:       "Run a SQL command for given arguments",
				Description: "Run a SQL command for given arguments",
				ArgsUsage:   "[name]",
				Action:      m.run,
				Flags: []cli.Flag{
					cli.StringSliceFlag{
						Name:  "param, p",
						Usage: "Parameters for the command",
					},
				},
			},
		},
	}
}

func (m *SQLCommand) beforeEach(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	m.dir = filepath.Join(dir, "/database/command")
	return nil
}

func (m *SQLCommand) create(ctx *cli.Context) error {
	args := ctx.Args()

	if len(args) != 1 {
		return cli.NewExitError("Create command expects a single argument", ErrCodeCommand)
	}

	generator := &script.Generator{
		Dir: m.dir,
	}

	container := strings.Replace(ctx.String("filename"), " ", "_", -1)
	name := strings.Replace(args[0], " ", "-", -1)
	path, err := generator.Create(container, name)

	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeCommand)
	}

	log.Infof("Command '%s' has been created at '%s' successfully", name, path)
	return nil
}

func (m *SQLCommand) run(ctx *cli.Context) error {
	args := ctx.Args()
	params := params(ctx.StringSlice("param"))

	if len(args) != 1 {
		return cli.NewExitError("Run command expects a single argument", ErrCodeCommand)
	}

	name := args[0]

	log.Infof("Running command '%s'", name)

	gateway, err := gateway(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if ioErr := gateway.Close(); err == nil {
			err = ioErr
		}
	}()

	runner := &script.Runner{
		Dir: m.dir,
		DB:  gateway.DB(),
	}

	rows := &gom.Rows{}
	rows, err = runner.Run(name, params...)

	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeCommand)
	}

	if err := printTable(rows); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeCommand)
	}

	log.Infof("Running command '%s' completed successfully", name)
	return nil
}

func printTable(rows *gom.Rows) error {
	table := tablewriter.NewWriter(os.Stdout)

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	table.SetHeader(columns)

	for rows.Next() {
		record, err := rows.SliceScan()
		if err != nil {
			return err
		}

		row := []string{}

		for _, column := range record {
			if data, ok := column.([]byte); ok {
				column = string(data)
			}
			row = append(row, fmt.Sprintf("%v", column))
		}

		table.Append(row)
	}

	table.Render()
	return nil
}

func params(args []string) []interface{} {
	result := []interface{}{}
	for _, arg := range args {
		result = append(result, arg)
	}
	return result
}
