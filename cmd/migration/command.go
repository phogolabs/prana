package migration

import (
	"github.com/apex/log"
	"github.com/urfave/cli"
)

const (
	ErrCodeMigration = 103
)

var Commands = []cli.Command{
	cli.Command{
		Name:        "setup",
		Usage:       "Setup the migration for the current project",
		Description: "Configures the current project by creating database directory hierarchy and initial migration",
		Action:      Setup,
	},
	cli.Command{
		Name:        "create",
		Usage:       "Generate a new migration with the given name, and the current timestamp as the version",
		Description: "Create a new migration file for the given name, and the current timestamp as the version in database/migration directory",
		ArgsUsage:   "[name]",
		Action:      Create,
	},
	cli.Command{
		Name:   "run",
		Usage:  "Runs the pending migrations",
		Action: Run,
	},
	cli.Command{
		Name:   "revert",
		Usage:  "Revert the latest applied migrations",
		Action: Revert,
	},
	cli.Command{
		Name:   "reset",
		Usage:  "Reverts and re-run all migrations",
		Action: Reset,
	},
	cli.Command{
		Name:   "status",
		Usage:  "Lists all migrations, marking those that have been applied",
		Action: Status,
	},
}

func Setup(ctx *cli.Context) error {
	migrator, err := Get(ctx)
	if err != nil {
		return err
	}

	if err := migrator.Setup(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	log.Infof("The project has been configured successfully")
	return nil
}

func Create(ctx *cli.Context) error {
	migrator, err := Get(ctx)
	if err != nil {
		return err
	}

	if len(ctx.Args()) != 1 {
		return cli.NewExitError("Create command expects a single argument", ErrCodeMigration)
	}

	path, err := migrator.Create(ctx.Args()[0])
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	log.Infof("Migration '%s' has been created successfully", path)
	return nil
}

func Run(ctx *cli.Context) error {
	return nil
}

func Revert(ctx *cli.Context) error {
	return nil
}

func Reset(ctx *cli.Context) error {
	return nil
}

func Status(ctx *cli.Context) error {
	return nil
}
