package main

import (
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/svett/gom/cmd"
	"github.com/svett/gom/cmd/migration"
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
		Name:   "database-driver",
		Value:  "postgres",
		Usage:  "Database Driver",
		EnvVar: "GOM_DB_DRIVER",
	},
	cli.StringFlag{
		Name:   "database-url",
		Value:  "",
		Usage:  "Database URL",
		EnvVar: "GOM_DB_URL",
	},
}

var commands = []cli.Command{
	cli.Command{
		Name:         "migration",
		Usage:        "A group of commands for generating, running, and reverting migrations",
		Description:  "A group of commands for generating, running, and reverting migrations",
		BashComplete: cli.DefaultAppComplete,
		Before:       migration.BeforeEach,
		Subcommands:  migration.Commands,
	},
}

func main() {
	app := &cli.App{
		Name:                 "gom",
		HelpName:             "gom",
		Usage:                "Golang Object Mapper",
		UsageText:            "gom [global options]",
		Version:              "0.1",
		BashComplete:         cli.DefaultAppComplete,
		EnableBashCompletion: true,
		Commands:             commands,
		Writer:               os.Stdout,
		ErrWriter:            os.Stderr,
		Flags:                flags,
		Before:               cmd.BeforeEach,
		After:                cmd.AfterEach,
	}

	app.Run(os.Args)
}
