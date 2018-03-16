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
		Name:   "setup",
		Usage:  "Setup the migration for the current project",
		Action: Setup,
	},
	cli.Command{
		Name:   "create",
		Usage:  "Create new migration",
		Action: Create,
	},
	cli.Command{
		Name:   "run",
		Usage:  "Apply the pending migrations",
		Action: Run,
	},
	cli.Command{
		Name:   "rollback",
		Usage:  "Rollback the applied migrations",
		Action: Rollback,
	},
	cli.Command{
		Name:   "status",
		Usage:  "Show the migration status",
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

func Rollback(ctx *cli.Context) error {
	return nil
}

func Status(ctx *cli.Context) error {
	return nil
}
