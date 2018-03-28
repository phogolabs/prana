package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
	"github.com/svett/gom"
	"github.com/urfave/cli"
)

const (
	ErrCodeArg       = 101
	ErrCodeMigration = 103
	ErrCodeCommand   = 104
)

type LogHandler struct {
	Writer io.Writer
}

func (h *LogHandler) HandleLog(entry *log.Entry) error {
	_, err := fmt.Fprintln(h.Writer, entry.Message)
	return err
}

func BeforeEach(ctx *cli.Context) error {
	var handler log.Handler

	if strings.EqualFold("json", ctx.String("log-format")) {
		handler = json.New(os.Stderr)
	} else {
		handler = &LogHandler{
			Writer: os.Stderr,
		}
	}

	log.SetHandler(handler)
	log.SetLevelFromString(ctx.String("log-level"))
	return nil
}

func Gateway(ctx *cli.Context) (*gom.Gateway, error) {
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
