package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/phogolabs/cli"
	"github.com/phogolabs/parcello"
	"github.com/phogolabs/prana/sqlmodel"
)

// SQLModel provides a subcommands to work generate structs from existing schema
type SQLModel struct {
	executor *sqlmodel.Executor
}

// CreateCommand creates a cli.Command that can be used by cli.App.
func (m *SQLModel) CreateCommand() *cli.Command {
	return &cli.Command{
		Name:        "model",
		Usage:       "A group of commands for generating object model from database schema",
		Description: "A group of commands for generating object model from database schema",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "package-dir, p",
				Usage: "path to the package, where the source code will be generated",
				Value: "./database/model",
			},
		},
		Commands: []*cli.Command{
			&cli.Command{
				Name:        "print",
				Usage:       "Print the object model for given database schema or tables",
				Description: "Print the object model for given database schema or tables",
				Action:      m.print,
				Before:      m.before,
				After:       m.after,
				Flags:       m.flags(),
			},
			&cli.Command{
				Name:        "sync",
				Usage:       "Generate a package of models for given database schema",
				Description: "Generate a package of models for given database schema",
				Action:      m.sync,
				Before:      m.before,
				After:       m.after,
				Flags:       m.flags(),
			},
		},
	}
}

func (m *SQLModel) flags() []cli.Flag {
	return []cli.Flag{
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
		&cli.StringFlag{
			Name:  "orm-tag, m",
			Usage: "tag tag that is wellknow for some ORM packages. supported: (sqlx, gorm)",
			Value: "sqlx",
		},
		&cli.StringSliceFlag{
			Name:  "extra-tag, e",
			Usage: "extra tags that should be included in model fields. supported: (json, xml, validate)",
		},
		&cli.BoolFlag{
			Name:  "include-docs, d",
			Usage: "include API documentation in generated source code",
			Value: true,
		},
	}
}

func (m *SQLModel) before(ctx *cli.Context) error {
	db, err := open(ctx)
	if err != nil {
		return err
	}

	provider, err := provider(db)
	if err != nil {
		return err
	}

	builder, err := m.builder(ctx)
	if err != nil {
		return err
	}

	m.executor = &sqlmodel.Executor{
		Provider: &sqlmodel.ModelProvider{
			Config: &sqlmodel.ModelProviderConfig{
				Package:        filepath.Base(ctx.GlobalString("package-dir")),
				UseNamedParams: ctx.Bool("use-named-params"),
				InlcudeDoc:     ctx.Bool("include-docs"),
			},
			TagBuilder: builder,
			Provider:   provider,
		},
		Generator: &sqlmodel.Codegen{
			Format: true,
		},
	}

	return nil
}

func (m *SQLModel) builder(ctx *cli.Context) (sqlmodel.TagBuilder, error) {
	registered := make(map[string]struct{})
	builder := sqlmodel.CompositeTagBuilder{}

	tags := []string{}
	tags = append(tags, ctx.String("orm-tag"))
	tags = append(tags, ctx.StringSlice("extra-tag")...)

	for _, tag := range tags {
		if _, ok := registered[tag]; ok {
			continue
		}

		registered[tag] = struct{}{}

		switch strings.ToLower(tag) {
		case "sqlx":
			builder = append(builder, sqlmodel.SQLXTagBuilder{})
		case "gorm":
			builder = append(builder, sqlmodel.GORMTagBuilder{})
		case "json":
			builder = append(builder, sqlmodel.JSONTagBuilder{})
		case "xml":
			builder = append(builder, sqlmodel.XMLTagBuilder{})
		case "validate":
			builder = append(builder, sqlmodel.ValidateTagBuilder{})
		default:
			err := fmt.Errorf("Cannot find tag builder for '%s'", tag)
			return nil, cli.NewExitError(err.Error(), ErrCodeArg)
		}
	}

	return builder, nil
}

func (m *SQLModel) after(ctx *cli.Context) error {
	if m.executor != nil {
		if err := m.executor.Provider.Close(); err != nil {
			return cli.NewExitError(err.Error(), ErrCodeSchema)
		}
	}
	return nil
}

func (m *SQLModel) print(ctx *cli.Context) error {
	if err := m.executor.Write(os.Stdout, m.spec(ctx)); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}

	return nil
}

func (m *SQLModel) sync(ctx *cli.Context) error {
	path, err := m.executor.Create(m.spec(ctx))
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeSchema)
	}

	if path != "" {
		log.Infof("Generated a database model at: '%s'", path)
	}

	return nil
}

func (m *SQLModel) spec(ctx *cli.Context) *sqlmodel.Spec {
	spec := &sqlmodel.Spec{
		Filename:     "schema.go",
		Template:     "model",
		FileSystem:   parcello.Dir(ctx.GlobalString("package-dir")),
		Schema:       ctx.String("schema-name"),
		Tables:       ctx.StringSlice("table-name"),
		IgnoreTables: ctx.StringSlice("ignore-table-name"),
	}

	return spec
}
