package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/jmoiron/sqlx"
	"github.com/phogolabs/prana/sqlmigr"
	"github.com/phogolabs/parcello"
	"github.com/urfave/cli"
)

// SQLMigration provides a subcommands to work with SQL migrations.
type SQLMigration struct {
	executor *sqlmigr.Executor
	db       *sqlx.DB
	cwd      string
}

// CreateCommand creates a cli.Command that can be used by cli.App.
func (m *SQLMigration) CreateCommand() cli.Command {
	return cli.Command{
		Name:         "migration",
		Usage:        "A group of commands for generating, running, and reverting migrations",
		Description:  "A group of commands for generating, running, and reverting migrations",
		BashComplete: cli.DefaultAppComplete,
		Before:       m.before,
		After:        m.after,
		Subcommands: []cli.Command{
			{
				Name:        "setup",
				Usage:       "Setup the migration for the current project",
				Description: "Configure the current project by creating database directory hierarchy and initial migration",
				Action:      m.setup,
			},
			{
				Name:        "create",
				Usage:       "Generate a new migration with the given name, and the current timestamp as the version",
				Description: "Create a new migration file for the given name, and the current timestamp as the version in database/migration directory",
				ArgsUsage:   "[name]",
				Action:      m.create,
			},
			{
				Name:   "run",
				Usage:  "Run the pending migrations",
				Action: m.run,
				Flags: []cli.Flag{
					cli.IntFlag{
						Name:  "count, c",
						Usage: "Number of migrations to be executed",
						Value: 1,
					},
				},
			},
			{
				Name:   "revert",
				Usage:  "Revert the latest applied migrations",
				Action: m.revert,
				Flags: []cli.Flag{
					cli.IntFlag{
						Name:  "count, c",
						Usage: "Number of migrations to be reverted",
						Value: 1,
					},
				},
			},
			{
				Name:   "reset",
				Usage:  "Revert and re-run all migrations",
				Action: m.reset,
			},
			{
				Name:   "status",
				Usage:  "Show all migrations, marking those that have been applied",
				Action: m.status,
			},
		},
	}
}

func (m *SQLMigration) before(ctx *cli.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	db, err := open(ctx)
	if err != nil {
		return err
	}

	fs := parcello.Dir(filepath.Join(cwd, "database/migration"))

	m.cwd = string(fs)
	m.db = db
	m.executor = &sqlmigr.Executor{
		Logger: log.Log,
		Provider: &sqlmigr.Provider{
			FileSystem: fs,
			DB:         db,
		},
		Runner: &sqlmigr.Runner{
			FileSystem: fs,
			DB:         db,
		},
		Generator: &sqlmigr.Generator{
			FileSystem: fs,
		},
	}

	return nil
}

func (m *SQLMigration) after(ctx *cli.Context) error {
	if err := m.db.Close(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	return nil
}

func (m *SQLMigration) setup(ctx *cli.Context) error {
	if err := m.executor.Setup(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	log.Infof("Setup project directory at: '%s'", m.cwd)
	return nil
}

func (m *SQLMigration) create(ctx *cli.Context) error {
	args := ctx.Args()

	if len(args) != 1 {
		return cli.NewExitError("Create command expects a single argument", ErrCodeMigration)
	}

	item, err := m.executor.Create(args[0])
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	log.Infof("Created migration at: '%s'", filepath.Join(m.cwd, item.Filename()))
	return nil
}

func (m *SQLMigration) run(ctx *cli.Context) error {
	count := ctx.Int("count")
	if count <= 0 {
		return cli.NewExitError("The count argument cannot be negative number", ErrCodeMigration)
	}

	_, err := m.executor.Run(count)
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	return nil
}

func (m *SQLMigration) revert(ctx *cli.Context) error {
	count := ctx.Int("count")
	if count <= 0 {
		return cli.NewExitError("The count argument cannot be negative number", ErrCodeMigration)
	}

	_, err := m.executor.Revert(count)
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	return nil
}

func (m *SQLMigration) reset(ctx *cli.Context) error {
	_, err := m.executor.RevertAll()
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	_, err = m.executor.RunAll()
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	return nil
}

func (m *SQLMigration) status(ctx *cli.Context) error {
	migrations, err := m.executor.Migrations()
	if err != nil {
		return err
	}

	if strings.EqualFold("json", ctx.GlobalString("log-format")) {
		sqlmigr.Flog(log.Log, migrations)
		return nil
	}

	sqlmigr.Ftable(os.Stdout, migrations)
	return nil
}
