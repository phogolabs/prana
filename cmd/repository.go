package cmd

import (
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/phogolabs/cli"
	"github.com/phogolabs/parcello"
	"github.com/phogolabs/prana/sqlmodel"
)

// SQLRepository provides a subcommands to work generate repository from existing schema
type SQLRepository struct {
	executor *sqlmodel.Executor
}

// CreateCommand creates a cli.Command that can be used by cli.App.
func (m *SQLRepository) CreateCommand() *cli.Command {
	return &cli.Command{
		Name:        "repository",
		Usage:       "A group of commands for generating database repository from schema",
		Description: "A group of commands for generating database repository from schema",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "package-dir, p",
				Usage: "path to the package, where the source code will be generated",
				Value: "./database",
			},
			&cli.StringFlag{
				Name:  "model-package-dir",
				Usage: "path to the model's package",
				Value: "./database/model",
			},
		},
		Commands: []*cli.Command{
			&cli.Command{
				Name:        "print",
				Usage:       "Print the database repositories for given database schema or tables",
				Description: "Print the database repositories for given database schema or tables",
				Action:      m.print,
				Before:      m.before,
				After:       m.after,
				Flags:       m.flags(false),
			},
			&cli.Command{
				Name:        "sync",
				Usage:       "Generate a package of repositories for given database schema",
				Description: "Generate a package of repositories for given database schema",
				Action:      m.sync,
				Before:      m.before,
				After:       m.after,
				Flags:       m.flags(true),
			},
		},
	}
}

func (m *SQLRepository) flags(include bool) []cli.Flag {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  "schema-name, s",
			Usage: "name of the database schema",
			Value: "",
		},
		&cli.StringSliceFlag{
			Name:  "table-name, t",
			Usage: "name of the table in the database",
		},
		&cli.StringSliceFlag{
			Name:  "ignore-table-name, i",
			Usage: "name of the table in the database that should be skipped",
			Value: []string{"migrations"},
		},
		&cli.BoolFlag{
			Name:  "include-docs, d",
			Usage: "include API documentation in generated source code",
			Value: true,
		},
	}

	if include {
		flag := &cli.BoolFlag{
			Name:  "include-tests",
			Usage: "include repository tests",
			Value: true,
		}

		flags = append(flags, flag)
	}

	return flags
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
				Package:        filepath.Base(ctx.GlobalString("model-package-dir")),
				UseNamedParams: ctx.Bool("use-named-params"),
				InlcudeDoc:     ctx.Bool("include-docs"),
			},
			TagBuilder: &sqlmodel.NoopTagBuilder{},
			Provider:   provider,
		},
		Generator: &sqlmodel.Codegen{
			Meta: map[string]interface{}{
				"RepositoryPackage": filepath.Base(ctx.GlobalString("package-dir")),
			},
			Format: true,
		},
	}

	return nil
}

func (m *SQLRepository) after(ctx *cli.Context) error {
	if m.executor != nil {
		if err := m.executor.Provider.Close(); err != nil {
			return cli.NewExitError(err.Error(), ErrCodeSchema)
		}
	}

	return nil
}

func (m *SQLRepository) print(ctx *cli.Context) error {
	for _, spec := range m.specs(ctx) {
		if err := m.executor.Write(os.Stdout, spec); err != nil {
			return cli.NewExitError(err.Error(), ErrCodeSchema)
		}
	}

	return nil
}

func (m *SQLRepository) sync(ctx *cli.Context) error {
	for _, spec := range m.specs(ctx) {
		path, err := m.executor.Create(spec)
		if err != nil {
			return cli.NewExitError(err.Error(), ErrCodeSchema)
		}

		if path != "" {
			log.Infof("Generated a database repository at: '%s'", path)
		}
	}

	return nil
}

func (m *SQLRepository) specs(ctx *cli.Context) []*sqlmodel.Spec {
	var (
		specs = []*sqlmodel.Spec{}
		spec  *sqlmodel.Spec
	)

	spec = &sqlmodel.Spec{
		Filename:     "repository.go",
		Template:     "repository",
		FileSystem:   parcello.Dir(ctx.GlobalString("package-dir")),
		Schema:       ctx.String("schema-name"),
		Tables:       ctx.StringSlice("table-name"),
		IgnoreTables: ctx.StringSlice("ignore-table-name"),
	}

	specs = append(specs, spec)

	if ctx.Bool("include-tests") {
		spec = &sqlmodel.Spec{
			Filename:     "repository_test.go",
			Template:     "repository_test",
			FileSystem:   spec.FileSystem,
			Schema:       spec.Schema,
			Tables:       spec.Tables,
			IgnoreTables: spec.IgnoreTables,
		}

		specs = append(specs, spec)

		spec = &sqlmodel.Spec{
			Filename:     "suite_test.go",
			Template:     "suite_test",
			FileSystem:   spec.FileSystem,
			Schema:       spec.Schema,
			Tables:       spec.Tables,
			IgnoreTables: spec.IgnoreTables,
		}

		specs = append(specs, spec)
	}

	return specs
}
