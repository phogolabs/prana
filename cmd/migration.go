package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/jmoiron/sqlx"
	"github.com/phogolabs/cli"
	"github.com/phogolabs/parcello"
	"github.com/phogolabs/prana/sqlmigr"
)

// SQLMigration provides a subcommands to work with SQL migrations.
type SQLMigration struct {
	executor *sqlmigr.Executor
	db       *sqlx.DB
	dir      string
}

// CreateCommand creates a cli.Command that can be used by cli.App.
func (m *SQLMigration) CreateCommand() *cli.Command {
	return &cli.Command{
		Name:        "migration",
		Usage:       "A group of commands for generating, running, and reverting migrations",
		Description: "A group of commands for generating, running, and reverting migrations",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "migration-dir, d",
				Usage:    "path to the directory that contain the migrations",
				EnvVar:   "PRANA_MIGRATION_DIR",
				Value:    "./database/migration",
				Required: true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "setup",
				Usage:       "Setup the migration for the current project",
				Description: "Configure the current project by creating database directory hierarchy and initial migration",
				Action:      m.setup,
				Before:      m.before,
				After:       m.after,
			},
			{
				Name:        "create",
				Usage:       "Generate a new migration with the given name, and the current timestamp as the version",
				Description: "Create a new migration file for the given name, and the current timestamp as the version in database/migration directory",
				ArgsUsage:   "[name]",
				Action:      m.create,
				Before:      m.before,
				After:       m.after,
			},
			{
				Name:   "run",
				Usage:  "Run the pending migrations",
				Action: m.run,
				Before: m.before,
				After:  m.after,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "count, c",
						Usage: "Number of migrations to be executed. Negative number will run all",
						Value: -1,
					},
				},
			},
			{
				Name:   "revert",
				Usage:  "Revert the latest applied migrations",
				Action: m.revert,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "count, c",
						Usage: "Number of migrations to be reverted. Negative number will revert all",
						Value: -1,
					},
				},
			},
			{
				Name:   "reset",
				Usage:  "Revert and re-run all migrations",
				Action: m.reset,
				Before: m.before,
				After:  m.after,
			},
			{
				Name:   "status",
				Usage:  "Show all migrations, marking those that have been applied",
				Action: m.status,
				Before: m.before,
				After:  m.after,
			},
		},
	}
}

func (m *SQLMigration) before(ctx *cli.Context) error {
	db, err := open(ctx)
	if err != nil {
		return err
	}

	m.dir, err = filepath.Abs(ctx.String("migration-dir"))
	if err != nil {
		return cli.WrapError(err, ErrCodeArg)
	}

	m.db = db
	m.executor = &sqlmigr.Executor{
		Logger: log.Log,
		Provider: &sqlmigr.Provider{
			FileSystem: parcello.Dir(m.dir),
			DB:         db,
		},
		Runner: &sqlmigr.Runner{
			FileSystem: parcello.Dir(m.dir),
			DB:         db,
		},
		Generator: &sqlmigr.Generator{
			FileSystem: parcello.Dir(m.dir),
		},
	}

	return nil
}

func (m *SQLMigration) after(ctx *cli.Context) error {
	if m.db != nil {
		if err := m.db.Close(); err != nil {
			return cli.NewExitError(err.Error(), ErrCodeMigration)
		}
	}

	return nil
}

func (m *SQLMigration) setup(ctx *cli.Context) error {
	if err := m.executor.Setup(); err != nil {
		if os.IsExist(err) {
			return nil
		}

		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	log.Infof("Setup project directory at: '%s'", m.dir)
	return nil
}

func (m *SQLMigration) create(ctx *cli.Context) error {
	args := ctx.Args

	if len(args) != 1 {
		return cli.NewExitError("Create command expects a single argument", ErrCodeMigration)
	}

	item, err := m.executor.Create(args[0])
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	log.Infof("Created migration at: '%s'", filepath.Join(m.dir, item.Filenames()[0]))
	return nil
}

func (m *SQLMigration) run(ctx *cli.Context) error {
	count := ctx.Int("count")

	_, err := m.executor.Run(count)
	if err != nil {
		err = m.errf(err)
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	return nil
}

func (m *SQLMigration) revert(ctx *cli.Context) error {
	count := ctx.Int("count")

	_, err := m.executor.Revert(count)
	if err != nil {
		err = m.errf(err)
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	return nil
}

func (m *SQLMigration) reset(ctx *cli.Context) error {
	_, err := m.executor.RevertAll()
	if err != nil {
		err = m.errf(err)
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	_, err = m.executor.RunAll()
	if err != nil {
		err = m.errf(err)
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

func (m *SQLMigration) errf(err error) error {
	if os.IsNotExist(err) {
		err = fmt.Errorf("Directory '%s' does not exist", m.dir)
	}
	return err
}
