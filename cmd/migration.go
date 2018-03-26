package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/svett/gom"
	"github.com/svett/gom/migration"
	"github.com/urfave/cli"
)

type Migration struct {
	executor *migration.Executor
	gateway  *gom.Gateway
}

func (m *Migration) CreateCommand() cli.Command {
	return cli.Command{
		Name:         "migration",
		Usage:        "A group of commands for generating, running, and reverting migrations",
		Description:  "A group of commands for generating, running, and reverting migrations",
		BashComplete: cli.DefaultAppComplete,
		Before:       m.BeforeEach,
		After:        m.AfterEach,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "setup",
				Usage:       "Setup the migration for the current project",
				Description: "Configures the current project by creating database directory hierarchy and initial migration",
				Action:      m.Setup,
			},
			cli.Command{
				Name:        "create",
				Usage:       "Generate a new migration with the given name, and the current timestamp as the version",
				Description: "Create a new migration file for the given name, and the current timestamp as the version in database/migration directory",
				ArgsUsage:   "[name]",
				Action:      m.Create,
			},
			cli.Command{
				Name:   "run",
				Usage:  "Runs the pending migrations",
				Action: m.Run,
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
				Action: m.Revert,
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
				Action: m.Reset,
			},
			cli.Command{
				Name:   "status",
				Usage:  "Lists all migrations, marking those that have been applied",
				Action: m.Status,
			},
		},
	}
}

func (m *Migration) BeforeEach(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	conn := ctx.GlobalString("database-url")

	uri, err := url.Parse(conn)
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	driver := uri.Scheme
	source := strings.Replace(conn, fmt.Sprintf("%s://", driver), "", -1)

	gateway, err := gom.Open(driver, source)
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	dir = filepath.Join(dir, "/database/migration")

	m.executor = &migration.Executor{
		Provider: &migration.Provider{
			Dir:     dir,
			Gateway: gateway,
		},
		Runner: &migration.Runner{
			Dir:     dir,
			Gateway: gateway,
		},
		Generator: &migration.Generator{
			Dir: dir,
		},
		OnRunFn:    m.OnRun,
		OnRevertFn: m.OnRevert,
	}

	return nil
}

func (m *Migration) AfterEach(ctx *cli.Context) error {
	if m.gateway == nil {
		return nil
	}

	if err := m.gateway.Close(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	m.gateway = nil
	return nil
}

func (m *Migration) Setup(ctx *cli.Context) error {
	if err := m.executor.Setup(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	log.Info("The project has been configured successfully")
	return nil
}

func (m *Migration) Create(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return cli.NewExitError("Create command expects a single argument", ErrCodeMigration)
	}

	name := ctx.Args()[0]
	name = strings.Replace(name, " ", "_", -1)

	path, err := m.executor.Create(name)
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	log.Infof("Migration '%s' has been created successfully", path)
	return nil
}

func (m *Migration) Run(ctx *cli.Context) error {
	count := ctx.Int("count")
	if count <= 0 {
		return cli.NewExitError("The count argument cannot be negative number", ErrCodeMigration)
	}

	if err := m.executor.Run(count); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	return nil
}

func (m *Migration) OnRun(item *migration.Item) {
	log.Infof("Running migration '%s'", item.Filename())
}

func (m *Migration) Revert(ctx *cli.Context) error {
	count := ctx.Int("count")
	if count <= 0 {
		return cli.NewExitError("The count argument cannot be negative number", ErrCodeMigration)
	}

	if err := m.executor.Revert(count); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	return nil
}

func (m *Migration) OnRevert(item *migration.Item) {
	log.Infof("Reverting migration '%s'", item.Filename())
}

func (m *Migration) Reset(ctx *cli.Context) error {
	const all = -1

	if err := m.executor.Revert(all); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	if err := m.executor.Run(all); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	return nil
}

func (m *Migration) Status(ctx *cli.Context) error {
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
