package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/jmoiron/sqlx"
	"github.com/phogolabs/oak/model"
	"github.com/urfave/cli"
)

// SQLModel provides a subcommands to work generate structs from existing schema
type SQLModel struct {
	db       *sqlx.DB
	executor *model.Executor
}

// CreateCommand creates a cli.Command that can be used by cli.App.
func (m *SQLModel) CreateCommand() cli.Command {
	return cli.Command{
		Name:         "model",
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
				Name:  "package-dir, p",
				Usage: "path to the package, where the source code will be generated",
				Value: "./database/model",
			},
			cli.StringFlag{
				Name:  "orm-type, m",
				Usage: "orm package for which the model tag will be generated",
				Value: "sqlx",
			},
			cli.BoolFlag{
				Name:  "keep-schema, k",
				Usage: "keep the schema as package (except default schema)",
			},
			cli.BoolTFlag{
				Name:  "include-docs, d",
				Usage: "include API documentation in generated source code",
			},
		},
		Subcommands: []cli.Command{
			{
				Name:        "print",
				Usage:       "Print the object model for given database schema or tables",
				Description: "Print the object model for given database schema or tables",
				Action:      m.print,
			},
			{
				Name:        "sync",
				Usage:       "Generate a package of models for given database schema",
				Description: "Generate a package of models for given database schema",
				Action:      m.sync,
			},
		},
	}
}

func (m *SQLModel) before(ctx *cli.Context) error {
	db, err := open(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}

	var provider model.Provider

	switch db.DriverName() {
	case "sqlite3":
		provider = &model.SQLiteProvider{DB: db}
	case "postgres":
		provider = &model.PostgreSQLProvider{DB: db}
	case "mysql":
		provider = &model.MySQLProvider{DB: db}
	default:
		err = fmt.Errorf("Cannot find provider for database driver '%s'", db.DriverName())
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}

	builder := model.CompositeTagBuilder{}

	switch strings.ToLower(ctx.String("orm-type")) {
	case "sqlx":
		builder = append(builder, model.SQLXTagBuilder{})
	case "gorm":
		builder = append(builder, model.GORMTagBuilder{})
	default:
		err = fmt.Errorf("Cannot find tag builder for '%s'", ctx.String("orm-type"))
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}

	builder = append(builder, model.JSONTagBuilder{})
	builder = append(builder, model.XMLTagBuilder{})

	m.db = db
	m.executor = &model.Executor{
		Config: &model.ExecutorConfig{
			KeepSchema: ctx.Bool("keep-schema"),
		},
		Provider: provider,
		Composer: &model.Generator{
			TagBuilder: builder,
			Config: &model.GeneratorConfig{
				KeepSchema:   ctx.Bool("keep-schema"),
				InlcudeDoc:   ctx.BoolT("include-docs"),
				IgnoreTables: ctx.StringSlice("ignore-table-name"),
			},
		},
	}

	return nil
}

func (m *SQLModel) after(ctx *cli.Context) error {
	if err := m.db.Close(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}
	return nil
}

func (m *SQLModel) print(ctx *cli.Context) error {
	spec := &model.Spec{
		Dir:    ctx.GlobalString("package-dir"),
		Schema: ctx.GlobalString("schema-name"),
		Tables: ctx.GlobalStringSlice("table-name"),
	}

	if err := m.executor.Write(os.Stdout, spec); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}

	return nil
}

func (m *SQLModel) sync(ctx *cli.Context) error {
	spec := &model.Spec{
		Dir:    ctx.GlobalString("package-dir"),
		Schema: ctx.GlobalString("schema-name"),
		Tables: ctx.GlobalStringSlice("table-name"),
	}

	path, err := m.executor.Create(spec)
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}

	if path != "" {
		log.Infof("Generated a database model at: '%s'", path)
	}

	return nil
}
