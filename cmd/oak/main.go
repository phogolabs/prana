// Command Line Interface of GOM.
package main

import (
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/phogolabs/oak/cmd"
	"github.com/urfave/cli"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name:   "log-level",
		Value:  "info",
		Usage:  "level of logging",
		EnvVar: "OAK_LOG_LEVEL",
	},
	cli.StringFlag{
		Name:   "log-format",
		Value:  "",
		Usage:  "format of the logs",
		EnvVar: "OAK_LOG_FORMAT",
	},
	cli.StringFlag{
		Name:   "database-url",
		Value:  "sqlite3://oak.db",
		Usage:  "Database URL",
		EnvVar: "OAK_DB_URL",
	},
}

func main() {
	migration := &cmd.SQLMigration{}
	script := &cmd.SQLScript{}
	schema := &cmd.SQLSchema{}

	commands := []cli.Command{
		migration.CreateCommand(),
		script.CreateCommand(),
		schema.CreateCommand(),
	}

	app := &cli.App{
		Name:                 "oak",
		HelpName:             "oak",
		Usage:                "Golang Database Object Manager",
		UsageText:            "oak [global options]",
		Version:              "1.0",
		BashComplete:         cli.DefaultAppComplete,
		EnableBashCompletion: true,
		Writer:               os.Stdout,
		ErrWriter:            os.Stderr,
		Flags:                flags,
		Before:               cmd.BeforeEach,
		Commands:             commands,
	}

	app.Run(os.Args)
}
