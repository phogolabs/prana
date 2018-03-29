// Package provides a set of commands used in CLI.
package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
	"github.com/phogolabs/gom"
	"github.com/urfave/cli"
)

const (
	// ErrCodeArg when the CLI argument is invalid.
	ErrCodeArg = 101
	// ErrCodeMigration when the migration operation fails.
	ErrCodeMigration = 103
	// ErrCodeCommand when the SQL command operation fails.
	ErrCodeCommand = 104
)

type logHandler struct {
	Writer io.Writer
}

func (h *logHandler) HandleLog(entry *log.Entry) error {
	_, err := fmt.Fprintln(h.Writer, entry.Message)
	return err
}

// BeforeEach is a function executed before each CLI operation.
func BeforeEach(ctx *cli.Context) error {
	var handler log.Handler

	if strings.EqualFold("json", ctx.String("log-format")) {
		handler = json.New(os.Stderr)
	} else {
		handler = &logHandler{
			Writer: os.Stderr,
		}
	}

	log.SetHandler(handler)
	log.SetLevelFromString(ctx.String("log-level"))
	return nil
}

func gateway(ctx *cli.Context) (*gom.Gateway, error) {
	conn := ctx.GlobalString("database-url")

	uri, err := url.Parse(conn)
	if err != nil {
		return nil, cli.NewExitError(err.Error(), ErrCodeArg)
	}

	driver := uri.Scheme
	source := strings.Replace(conn, fmt.Sprintf("%s://", driver), "", -1)

	gateway, err := gom.Open(driver, source)
	if err != nil {
		return nil, cli.NewExitError(err.Error(), ErrCodeArg)
	}

	return gateway, nil
}
