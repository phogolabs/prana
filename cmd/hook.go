package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
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
