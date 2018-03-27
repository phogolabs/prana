package main

import (
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/svett/gom/cmd"
	"github.com/urfave/cli"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name:   "log-level",
		Value:  "info",
		Usage:  "level of logging",
		EnvVar: "GOM_LOG_LEVEL",
	},
	cli.StringFlag{
		Name:   "log-format",
		Value:  "",
		Usage:  "format of the logs",
		EnvVar: "GOM_LOG_FORMAT",
	},
	cli.StringFlag{
		Name:   "database-url",
		Value:  "",
		Usage:  "Database URL",
		EnvVar: "GOM_DB_URL",
	},
}

func main() {
	migration := &cmd.SQLMigration{}
	command := &cmd.SQLCommand{}

	commands := []cli.Command{
		migration.CreateCommand(),
		command.CreateCommand(),
	}

	app := &cli.App{
		Name:                 "gom",
		HelpName:             "gom",
		Usage:                "Golang Object Mapper",
		UsageText:            "gom [global options]",
		Version:              "0.1",
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
