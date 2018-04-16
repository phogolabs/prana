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
	"github.com/phogolabs/oak/migration"
	"github.com/phogolabs/parcel"
	"github.com/urfave/cli"
)

// SQLMigration provides a subcommands to work with SQL migrations.
type SQLMigration struct {
	executor *migration.Executor
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
			cli.Command{
				Name:        "setup",
				Usage:       "Setup the migration for the current project",
				Description: "Configure the current project by creating database directory hierarchy and initial migration",
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
				Usage:  "Revert and re-run all migrations",
				Action: m.reset,
			},
			cli.Command{
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

	fs := parcel.Dir(filepath.Join(cwd, "database/migration"))

	m.cwd = string(fs)
	m.db = db
	m.executor = &migration.Executor{
		Logger: log.Log,
		Provider: &migration.Provider{
			FileSystem: fs,
			DB:         db,
		},
		Runner: &migration.Runner{
			FileSystem: fs,
			DB:         db,
		},
		Generator: &migration.Generator{
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
		m.log(migrations)
		return nil
	}

	m.table(migrations)
	return nil
}

func (m *SQLMigration) log(migrations []migration.Item) {
	for _, m := range migrations {
		status := "pending"
		timestamp := ""

		if !m.CreatedAt.IsZero() {
			status = "executed"
			timestamp = m.CreatedAt.Format(time.UnixDate)
		}

		fields := log.Fields{
			"Id":          m.ID,
			"Description": m.Description,
			"Status":      status,
			"CreatedAt":   timestamp,
		}

		log.WithFields(fields).Info("Migration")
	}
}

func (m *SQLMigration) table(migrations []migration.Item) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Description", "Status", "Created At"})

	for _, m := range migrations {
		status := color.YellowString("pending")
		timestamp := ""

		if !m.CreatedAt.IsZero() {
			status = color.GreenString("executed")
			timestamp = m.CreatedAt.Format(time.UnixDate)
		}

		row := []string{m.ID, m.Description, status, timestamp}
		table.Append(row)
	}

	table.Render()
}
