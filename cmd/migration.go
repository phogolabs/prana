package cmd

import (
	"os"
	"path/filepath"
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

func (m *Migration) Command() cli.Command {
	commands := []cli.Command{
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
					Name:  "step",
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
					Name:  "step",
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
	}

	return cli.Command{
		Name:         "migration",
		Usage:        "A group of commands for generating, running, and reverting migrations",
		Description:  "A group of commands for generating, running, and reverting migrations",
		BashComplete: cli.DefaultAppComplete,
		Before:       m.BeforeEach,
		After:        m.AfterEach,
		Subcommands:  commands,
	}
}

func (m *Migration) BeforeEach(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	gateway, err := gom.Open(ctx.GlobalString("database-driver"), ctx.GlobalString("database-url"))
	if err != nil {
		return err
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
	}

	return nil
}

func (m *Migration) AfterEach(ctx *cli.Context) error {
	if m.gateway == nil {
		return nil
	}

	if err := m.gateway.Close(); err != nil {
		return err
	}

	m.gateway = nil
	return nil
}

func (m *Migration) Setup(ctx *cli.Context) error {
	if err := m.executor.Setup(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	log.Infof("The project has been configured successfully")
	return nil
}

func (m *Migration) Create(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return cli.NewExitError("Create command expects a single argument", ErrCodeMigration)
	}

	path, err := m.executor.Create(ctx.Args()[0])
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	log.Infof("Migration '%s' has been created successfully", path)
	return nil
}

func (m *Migration) Run(ctx *cli.Context) error {
	if err := m.executor.Run(ctx.Int("step")); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	return nil
}

func (m *Migration) Revert(ctx *cli.Context) error {
	if err := m.executor.Revert(ctx.Int("step")); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	return nil
}

func (m *Migration) Reset(ctx *cli.Context) error {
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
