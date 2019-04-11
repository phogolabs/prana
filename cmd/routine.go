package cmd

import (
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/phogolabs/parcello"
	"github.com/phogolabs/prana/sqlexec"
	"github.com/phogolabs/prana/sqlmodel"
	"github.com/urfave/cli"
)

// SQLRoutine provides a subcommands to work with SQL scripts and their
// statements.
type SQLRoutine struct {
	runner   *sqlexec.Runner
	executor *sqlmodel.Executor
}

// CreateCommand creates a cli.Command that can be used by cli.App.
func (m *SQLRoutine) CreateCommand() cli.Command {
	return cli.Command{
		Name:         "routine",
		Usage:        "A group of commands for generating, running, and removing SQL commands",
		Description:  "A group of commands for generating, running, and removing SQL commands",
		BashComplete: cli.DefaultAppComplete,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "routine-dir, d",
				Usage:  "path to the directory that contain the SQL routines",
				EnvVar: "PRANA_ROUTINE_DIR",
				Value:  "./database/routine",
			},
		},
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "sync",
				Usage:       "Generate a SQL script of CRUD operations for given database schema",
				Description: "Generate a SQL script of CRUD operations for given database schema",
				Action:      m.sync,
				Before:      m.before,
				After:       m.after,
				Flags:       m.flags(),
			},
			cli.Command{
				Name:        "print",
				Usage:       "Print a SQL script of CRUD operations for given database schema",
				Description: "Print a SQL script of CRUD operations for given database schema",
				Action:      m.print,
				Before:      m.before,
				After:       m.after,
				Flags:       m.flags(),
			},
			cli.Command{
				Name:        "create",
				Usage:       "Create a new SQL command for given container filename",
				Description: "Create a new SQL command for given container filename",
				ArgsUsage:   "[name]",
				Action:      m.create,
				Before:      m.before,
				After:       m.after,
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
				Before:      m.before,
				After:       m.after,
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

func (m *SQLRoutine) flags() []cli.Flag {
	return []cli.Flag{
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
		cli.BoolTFlag{
			Name:  "use-named-params, n",
			Usage: "use named parameter instead of questionmark",
		},
		cli.BoolTFlag{
			Name:  "include-docs, d",
			Usage: "include API documentation in generated source code",
		},
	}
}

func (m *SQLRoutine) before(ctx *cli.Context) error {
	dir, err := filepath.Abs(ctx.GlobalString("routine-dir"))
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}

	db, err := open(ctx)
	if err != nil {
		return err
	}

	provider, err := provider(db)
	if err != nil {
		return err
	}

	m.runner = &sqlexec.Runner{
		FileSystem: parcello.Dir(dir),
		DB:         db,
	}

	m.executor = &sqlmodel.Executor{
		Provider: &sqlmodel.ModelProvider{
			Config: &sqlmodel.ModelProviderConfig{
				Package:        filepath.Base(ctx.GlobalString("routine-dir")),
				UseNamedParams: ctx.BoolT("use-named-params"),
				InlcudeDoc:     ctx.BoolT("include-docs"),
			},
			TagBuilder: &sqlmodel.NoopTagBuilder{},
			Provider:   provider,
		},
		Generator: &sqlmodel.Codegen{
			Format: false,
		},
	}

	return nil
}

func (m *SQLRoutine) create(ctx *cli.Context) error {
	args := ctx.Args()

	if len(args) != 1 {
		return cli.NewExitError("Create command expects a single argument", ErrCodeCommand)
	}

	generator := &sqlexec.Generator{
		FileSystem: m.runner.FileSystem,
	}

	name, path, err := generator.Create(ctx.String("filename"), args[0])
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeCommand)
	}

	dir, err := filepath.Abs(ctx.GlobalString("routine-dir"))
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}

	log.Infof("Created command '%s' at '%s'", name, filepath.Join(dir, path))
	return nil
}

func (m *SQLRoutine) run(ctx *cli.Context) error {
	args := ctx.Args()
	params := params(ctx.StringSlice("param"))

	if len(args) != 1 {
		return cli.NewExitError("Run command expects a single argument", ErrCodeCommand)
	}

	name := args[0]
	log.Infof("Running command '%s' from '%v'", name, m.runner.FileSystem)

	rows, err := m.runner.Run(name, params...)
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeCommand)
	}

	if err := m.runner.Print(os.Stdout, rows); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeCommand)
	}

	return nil
}

func (m *SQLRoutine) after(ctx *cli.Context) error {
	if m.executor == nil {
		return nil
	}

	if err := m.executor.Provider.Close(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}
	return nil
}

func (m *SQLRoutine) print(ctx *cli.Context) error {
	if err := m.executor.Write(os.Stdout, m.spec(ctx)); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}

	return nil
}

func (m *SQLRoutine) sync(ctx *cli.Context) error {
	path, err := m.executor.Create(m.spec(ctx))
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}

	if path != "" {
		log.Infof("Generated a database model at: '%s'", path)
	}

	return nil
}

func (m *SQLRoutine) spec(ctx *cli.Context) *sqlmodel.Spec {
	spec := &sqlmodel.Spec{
		Filename:     "routine.sql",
		Template:     "routine",
		FileSystem:   parcello.Dir(ctx.GlobalString("routine-dir")),
		Schema:       ctx.String("schema-name"),
		Tables:       ctx.StringSlice("table-name"),
		IgnoreTables: ctx.StringSlice("ignore-table-name"),
	}

	return spec
}

func params(args []string) []interface{} {
	result := []interface{}{}
	for _, arg := range args {
		result = append(result, arg)
	}
	return result
}
