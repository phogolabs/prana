package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
	"github.com/olekukonko/tablewriter"
	"github.com/phogolabs/gom/migration"
	"github.com/urfave/cli"
)

// SQLCommand provides a subcommands to work with SQL migrations.
type SQLMigration struct {
	executor *migration.Executor
	db       *sqlx.DB
	dir      string
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
			cli.Command{
				Name:        "setup",
				Usage:       "Setup the migration for the current project",
				Description: "Configures the current project by creating database directory hierarchy and initial migration",
				Action:      m.setup,
			},
			cli.Command{
				Name:        "create",
				Usage:       "Generate a new migration with the given name, and the current timestamp as the version",
				Description: "Create a new migration file for the given name, and the current timestamp as the version in database/migration directory",
				ArgsUsage:   "[name]",
				Action:      m.create,
			},
			cli.Command{
				Name:   "run",
				Usage:  "Runs the pending migrations",
				Action: m.run,
				Flags: []cli.Flag{
					cli.IntFlag{
						Name:  "count, c",
						Usage: "Number of migrations to be executed",
						Value: 1,
					},
				},
			},
			cli.Command{
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
			cli.Command{
				Name:   "reset",
				Usage:  "Reverts and re-run all migrations",
				Action: m.reset,
			},
			cli.Command{
				Name:   "status",
				Usage:  "Lists all migrations, marking those that have been applied",
				Action: m.status,
			},
		},
	}
}

func (m *SQLMigration) before(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	dir = filepath.Join(dir, "/database/migration")
	db, err := open(ctx)
	if err != nil {
		return err
	}

	m.dir = dir
	m.db = db
	m.executor = &migration.Executor{
		Provider: &migration.Provider{
			Dir: dir,
			DB:  db,
		},
		Runner: &migration.Runner{
			Dir: dir,
			DB:  db,
		},
		Generator: &migration.Generator{
			Dir: dir,
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

	log.Infof("Setup project directory at: '%s'", m.dir)
	return nil
}

func (m *SQLMigration) create(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return cli.NewExitError("Create command expects a single argument", ErrCodeMigration)
	}

	name := ctx.Args()[0]
	name = strings.Replace(name, " ", "_", -1)

	path, err := m.executor.Create(name)
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	log.Infof("Created migration at: '%s'", path)
	return nil
}

func (m *SQLMigration) run(ctx *cli.Context) error {
	count := ctx.Int("count")
	if count <= 0 {
		return cli.NewExitError("The count argument cannot be negative number", ErrCodeMigration)
	}

	log.Infof("Running %d pending migration(s)", count)
	if err := m.executor.Run(count); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	return nil
}

func (m *SQLMigration) revert(ctx *cli.Context) error {
	count := ctx.Int("count")
	if count <= 0 {
		return cli.NewExitError("The count argument cannot be negative number", ErrCodeMigration)
	}

	log.Infof("Reverting %d pending migration(s)", count)
	if err := m.executor.Revert(count); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	return nil
}

func (m *SQLMigration) reset(ctx *cli.Context) error {
	log.Info("Reverting all migrations")
	if err := m.executor.RevertAll(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	log.Info("Running all pending migrations")
	if err := m.executor.RunAll(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	return nil
}

func (m *SQLMigration) status(ctx *cli.Context) error {
	migrations, err := m.executor.Migrations()
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Description", "Status", "Created At"})

	for _, m := range migrations {
		status := color.YellowString("pending")
		timestamp := ""

		if !m.CreatedAt.IsZero() {
			status = color.GreenString("executed")
			timestamp = m.CreatedAt.Format(time.UnixDate)
		}

		row := []string{m.Id, m.Description, status, timestamp}
		table.Append(row)
	}

	table.Render()
	return nil
}
