// Package cmd provides a set of commands used in CLI.
package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
	"github.com/jmoiron/sqlx"
	"github.com/phogolabs/cli"
	"github.com/phogolabs/prana"
	"github.com/phogolabs/prana/sqlmodel"
)

const (
	// ErrCodeArg when the CLI argument is invalid.
	ErrCodeArg = 101
	// ErrCodeMigration when the migration operation fails.
	ErrCodeMigration = 103
	// ErrCodeCommand when the SQL command operation fails.
	ErrCodeCommand = 104
	// ErrCodeSchema when the SQL schema operation fails.
	ErrCodeSchema = 105
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

func open(ctx *cli.Context) (*sqlx.DB, error) {
	driver, conn, err := prana.ParseURL(ctx.GlobalString("database-url"))
	if err != nil {
		return nil, cli.NewExitError(err.Error(), ErrCodeArg)
	}

	db, err := sqlx.Open(driver, conn)
	if err != nil {
		return nil, cli.NewExitError(err.Error(), ErrCodeArg)
	}

	return db, nil
}

func provider(db *sqlx.DB) (sqlmodel.SchemaProvider, error) {
	switch db.DriverName() {
	case "sqlite3":
		return &sqlmodel.SQLiteProvider{DB: db}, nil
	case "postgres":
		return &sqlmodel.PostgreSQLProvider{DB: db}, nil
	case "mysql":
		return &sqlmodel.MySQLProvider{DB: db}, nil
	default:
		err := fmt.Errorf("Cannot find provider for database driver '%s'", db.DriverName())
		return nil, cli.NewExitError(err.Error(), ErrCodeArg)
	}
}
