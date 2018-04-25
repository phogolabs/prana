package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/jmoiron/sqlx"
	"github.com/olekukonko/tablewriter"
	"github.com/phogolabs/parcello"
	"github.com/phogolabs/prana/sqlexec"
	"github.com/urfave/cli"
)

// SQLScript provides a subcommands to work with SQL scripts and their
// statements.
type SQLScript struct {
	dir string
}

// CreateCommand creates a cli.Command that can be used by cli.App.
func (m *SQLScript) CreateCommand() cli.Command {
	return cli.Command{
		Name:         "script",
		Usage:        "A group of commands for generating, running, and removing SQL commands",
		Description:  "A group of commands for generating, running, and removing SQL commands",
		BashComplete: cli.DefaultAppComplete,
		Before:       m.before,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "directory, dir, d",
				Usage: "path to the directory that contain the scripts",
				Value: "./database/script",
			},
		},
		Subcommands: []cli.Command{
			{
				Name:        "sync",
				Usage:       "Generate a SQL script of CRUD operations for given database schema",
				Description: "Generate a SQL script of CRUD operations for given database schema",
				Action:      m.sync,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "schema-name, s",
						Usage: "name of the database schema",
						Value: "",
					},
					cli.StringSliceFlag{
						Name:  "table-name, t",
						Usage: "name of the table in the database",
					},
					cli.StringSliceFlag{
						Name:  "ignore-table-name, i",
						Usage: "name of the table in the database that should be skipped",
						Value: &cli.StringSlice{"migrations"},
					},
					cli.BoolFlag{
						Name:  "use-named-params, n",
						Usage: "use named parameter instead of questionmark",
					},
					cli.BoolTFlag{
						Name:  "include-docs, d",
						Usage: "include API documentation in generated source code",
					},
				},
			},
			{
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
			{
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

func (m *SQLScript) before(ctx *cli.Context) error {
	var err error
	m.dir, err = filepath.Abs(ctx.String("directory"))
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}

	return nil
}

func (m *SQLScript) create(ctx *cli.Context) error {
	args := ctx.Args()

	if len(args) != 1 {
		return cli.NewExitError("Create command expects a single argument", ErrCodeCommand)
	}

	generator := &sqlexec.Generator{
		FileSystem: parcello.Dir(m.dir),
	}

	name, path, err := generator.Create(ctx.String("filename"), args[0])
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeCommand)
	}

	log.Infof("Created command '%s' at '%s'", name, filepath.Join(m.dir, path))
	return nil
}

func (m *SQLScript) run(ctx *cli.Context) error {
	args := ctx.Args()
	params := params(ctx.StringSlice("param"))

	if len(args) != 1 {
		return cli.NewExitError("Run command expects a single argument", ErrCodeCommand)
	}

	name := args[0]

	log.Infof("Running command '%s' from '%s'", name, m.dir)

	db, err := open(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if ioErr := db.Close(); err == nil {
			err = ioErr
		}
	}()

	runner := &sqlexec.Runner{
		FileSystem: parcello.Dir(m.dir),
		DB:         db,
	}

	var rows *sqlx.Rows
	rows, err = runner.Run(name, params...)

	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeCommand)
	}

	if err := m.print(rows); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeCommand)
	}

	return nil
}

func (m *SQLScript) print(rows *sqlx.Rows) error {
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

func (m *SQLScript) sync(ctx *cli.Context) error {
	model := &SQLModel{skip: true}

	if err := model.before(ctx); err != nil {
		return err
	}

	if err := model.script(ctx); err != nil {
		_ = model.after(ctx)
		return err
	}

	return model.after(ctx)
}

func params(args []string) []interface{} {
	result := []interface{}{}
	for _, arg := range args {
		result = append(result, arg)
	}
	return result
}
