package cmd

import (
	"os"
	"strings"

	"github.com/apex/log"
	clix "github.com/apex/log/handlers/cli"
	jsonx "github.com/apex/log/handlers/json"
	"github.com/svett/gom"
	"github.com/urfave/cli"
)

const (
	ErrCodeArg = 101
	ErrCodeDb  = 102
)

func BeforeEach(ctx *cli.Context) error {
	var handler log.Handler

	if strings.EqualFold("json", ctx.String("log-format")) {
		handler = jsonx.New(os.Stderr)
	} else {
		handler = clix.New(os.Stderr)
	}

	log.SetHandler(handler)
	log.SetLevelFromString(ctx.String("log-level"))

	log.Log = log.Log.WithFields(
		log.Fields{
			"app_name":    ctx.App.Name,
			"app_version": ctx.App.Version,
		},
	)

	driver := ctx.String("database-driver")

	if driver == "" {
		return cli.NewExitError("Database driver is not initialized", ErrCodeArg)
	}

	source := ctx.String("database-url")
	if source == "" {
		return cli.NewExitError("Database source is not initialized", ErrCodeArg)
	}

	gateway, err := gom.Open(driver, source)
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeDb)
	}

	if err := gateway.DB().Ping(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeDb)
	}

	ctx.App.Metadata["gateway"] = gateway
	return nil
}

func AfterEach(ctx *cli.Context) error {
	metadata := ctx.App.Metadata
	gateway, ok := metadata["gateway"].(*gom.Gateway)
	if !ok {
		return nil
	}

	if err := gateway.Close(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeDb)
	}
	return nil
}
