package cmd

import (
	"fmt"
	"os"

	"github.com/apex/log"
	"github.com/jmoiron/sqlx"
	"github.com/phogolabs/gom/schema"
	"github.com/urfave/cli"
)

// SQLSchema provides a subcommands to work generate structs from existing schema
type SQLSchema struct {
	db       *sqlx.DB
	executor *schema.Executor
}

// CreateCommand creates a cli.Command that can be used by cli.App.
func (m *SQLSchema) CreateCommand() cli.Command {
	return cli.Command{
		Name:         "schema",
		Usage:        "A group of commands for generating object model from database schema",
		Description:  "A group of commands for generating object model from database schema",
		BashComplete: cli.DefaultAppComplete,
		Before:       m.before,
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
			cli.StringFlag{
				Name:  "schema-dir, d",
				Usage: "directory path to the package, for which the source code will be generated. the basename will be used as package name for default database schema.",
				Value: "./database/model",
			},
			cli.StringSliceFlag{
				Name:  "tag, g",
				Usage: "tag name that will be added to each of the struct's fields",
			},
			cli.BoolTFlag{
				Name:  "include-docs, D",
				Usage: "include API documentation in generated source code",
			},
		},
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "print",
				Usage:       "Print the object model for given database schema or tables",
				Description: "Print the object model for given database schema or tables",
				Action:      m.print,
			},
			cli.Command{
				Name:        "sync",
				Usage:       "Generate a package of models for given database schema",
				Description: "Generate a package of models for given database schema",
				Action:      m.sync,
			},
		},
	}
}

func (m *SQLSchema) before(ctx *cli.Context) error {
	db, err := open(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}

	var provider schema.Provider

	switch db.DriverName() {
	case "sqlite3":
		provider = &schema.SQLiteProvider{DB: db}
	case "postgres":
		provider = &schema.PostgreSQLProvider{DB: db}
	case "mysql":
		provider = &schema.MySQLProvider{DB: db}
	default:
		err = fmt.Errorf("Cannot find provider for database driver '%s'", db.DriverName())
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}

	m.db = db
	m.executor = &schema.Executor{
		Provider: provider,
		Composer: &schema.Generator{
			Config: &schema.GeneratorConfig{
				InlcudeDoc:   ctx.BoolT("include-docs"),
				IgnoreTables: ctx.StringSlice("ignore-table-name"),
			},
		},
	}

	return nil
}

func (m *SQLSchema) after(ctx *cli.Context) error {
	if err := m.db.Close(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}
	return nil
}

func (m *SQLSchema) print(ctx *cli.Context) error {
	spec := &schema.Spec{
		Dir:    ctx.GlobalString("schema-dir"),
		Schema: ctx.GlobalString("schema-name"),
		Tables: ctx.GlobalStringSlice("table-name"),
	}

	if err := m.executor.WriteTo(os.Stdout, spec); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}

	return nil
}

func (m *SQLSchema) sync(ctx *cli.Context) error {
	spec := &schema.Spec{
		Dir:    ctx.GlobalString("schema-dir"),
		Schema: ctx.GlobalString("schema-name"),
		Tables: ctx.GlobalStringSlice("table-name"),
	}

	path, err := m.executor.Create(spec)
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}

	log.Infof("Generated a schema model at: '%s'", path)
	return nil
}
