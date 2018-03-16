package cmd

import (
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/apex/log/handlers/json"
	"github.com/svett/gom"
	. "github.com/urfave/cli"
)

const (
	ErrCodeArg = 101
	ErrCodeDb  = 102
)

func BeforeEach(ctx *Context) error {
	var handler log.Handler

	if strings.EqualFold("json", ctx.String("log-format")) {
		handler = json.New(os.Stderr)
	} else {
		handler = cli.New(os.Stderr)
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
		return NewExitError("Database driver is not initialized", ErrCodeArg)
	}

	source := ctx.String("database-url")
	if source == "" {
		return NewExitError("Database source is not initialized", ErrCodeArg)
	}

	gateway, err := gom.Open(driver, source)
	if err != nil {
		return NewExitError(err.Error(), ErrCodeDb)
	}

	if err := gateway.DB().Ping(); err != nil {
		return NewExitError(err.Error(), ErrCodeDb)
	}

	ctx.App.Metadata["gateway"] = gateway
	return nil
}

func AfterEach(ctx *Context) error {
	metadata := ctx.App.Metadata
	gateway, ok := metadata["gateway"].(*gom.Gateway)
	if !ok {
		return nil
	}

	if err := gateway.Close(); err != nil {
		return NewExitError(err.Error(), ErrCodeDb)
	}
	return nil
}
