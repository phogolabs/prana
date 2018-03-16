package migration

import (
	"os"

	"github.com/svett/gom"
	"github.com/urfave/cli"
)

func BeforeEach(ctx *cli.Context) error {
	metadata := ctx.App.Metadata
	gateway, ok := metadata["gateway"].(*gom.Gateway)
	if !ok {
		return cli.NewExitError("Database connection is not established", ErrCodeMigration)
	}

	dir, err := os.Getwd()
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeMigration)
	}

	migrator := &gom.Migrator{
		Dir:     dir,
		Gateway: gateway,
	}
	ctx.App.Metadata["migrator"] = migrator
	return nil
}

func Get(ctx *cli.Context) (*gom.Migrator, error) {
	metadata := ctx.App.Metadata
	migrator, ok := metadata["migrator"].(*gom.Migrator)
	if !ok {
		return nil, cli.NewExitError("Migrator is not initialized", ErrCodeMigration)
	}
	return migrator, nil
}
