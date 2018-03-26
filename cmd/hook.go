package cmd

import (
	"os"
	"strings"

	"github.com/apex/log"
	clix "github.com/apex/log/handlers/cli"
	jsonx "github.com/apex/log/handlers/json"
	"github.com/urfave/cli"
)

const (
	ErrCodeArg       = 101
	ErrCodeMigration = 103
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

	return nil
}
