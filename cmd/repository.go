package cmd

import (
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/phogolabs/parcello"
	"github.com/phogolabs/prana/sqlmodel"
	"github.com/urfave/cli"
)

// SQLRepository provides a subcommands to work generate repository from existing schema
type SQLRepository struct {
	executor *sqlmodel.Executor
}

// CreateCommand creates a cli.Command that can be used by cli.App.
func (m *SQLRepository) CreateCommand() cli.Command {
	return cli.Command{
		Name:         "repository",
		Usage:        "A group of commands for generating database repository from schema",
		Description:  "A group of commands for generating database repository from schema",
		BashComplete: cli.DefaultAppComplete,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "package-dir, p",
				Usage: "path to the package, where the source code will be generated",
				Value: "./database",
			},
		},
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "print",
				Usage:       "Print the database repositories for given database schema or tables",
				Description: "Print the database repositories for given database schema or tables",
				Action:      m.print,
				Before:      m.before,
				After:       m.after,
				Flags:       m.flags(),
			},
			cli.Command{
				Name:        "sync",
				Usage:       "Generate a package of repositories for given database schema",
				Description: "Generate a package of repositories for given database schema",
				Action:      m.sync,
				Before:      m.before,
				After:       m.after,
				Flags:       m.flags(),
			},
		},
	}
}

func (m *SQLRepository) flags() []cli.Flag {
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
			Name:  "include-docs, d",
			Usage: "include API documentation in generated source code",
		},
	}
}

func (m *SQLRepository) before(ctx *cli.Context) error {
	db, err := open(ctx)
	if err != nil {
		return err
	}

	provider, err := provider(db)
	if err != nil {
		return err
	}

	m.executor = &sqlmodel.Executor{
		Provider: &sqlmodel.ModelProvider{
			Config: &sqlmodel.ModelProviderConfig{
				Package:        filepath.Base(ctx.GlobalString("package-dir")),
				UseNamedParams: ctx.Bool("use-named-params"),
				InlcudeDoc:     ctx.BoolT("include-docs"),
			},
			TagBuilder: &sqlmodel.NoopTagBuilder{},
			Provider:   provider,
		},
		Generator: &sqlmodel.Codegen{
			Format:   true,
			Template: "repository",
		},
	}

	return nil
}

func (m *SQLRepository) after(ctx *cli.Context) error {
	if err := m.executor.Provider.Close(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}
	return nil
}

func (m *SQLRepository) print(ctx *cli.Context) error {
	if err := m.executor.Write(os.Stdout, m.spec(ctx)); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}

	return nil
}

func (m *SQLRepository) sync(ctx *cli.Context) error {
	path, err := m.executor.Create(m.spec(ctx))
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}

	if path != "" {
		log.Infof("Generated a database repository at: '%s'", path)
	}

	return nil
}

func (m *SQLRepository) spec(ctx *cli.Context) *sqlmodel.Spec {
	spec := &sqlmodel.Spec{
		Filename:     "repository.go",
		FileSystem:   parcello.Dir(ctx.GlobalString("package-dir")),
		Schema:       ctx.String("schema-name"),
		Tables:       ctx.StringSlice("table-name"),
		IgnoreTables: ctx.StringSlice("ignore-table-name"),
	}

	return spec
}
