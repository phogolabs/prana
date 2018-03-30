// GOM package provides a wrapper to work with loukoum built queries as well
// maitaining database version by creating, executing and reverting SQL
// migrations.
//
// The package allows executing embedded SQL statements from script for a given
// name.
package gom

import (
	"database/sql"
	"io"

	"github.com/jmoiron/sqlx"
	"github.com/phogolabs/gom/script"
)

// Entity is a destination object for given select operation.
type Entity = interface{}

// Rows is a wrapper around sql.Rows which caches costly reflect operations
// during a looped StructScan
type Rows = sqlx.Rows

// Row is a reimplementation of sql.Row in order to gain access to the underlying
// sql.Rows.Columns() data, necessary for StructScan.
type Row = sqlx.Row

// A Result summarizes an executed SQL command.
type Result = sql.Result

// Tx is an sqlx wrapper around sql.Tx with extra functionality
type Tx = sqlx.Tx

var provider *script.Provider

func init() {
	provider = &script.Provider{}
}

// Load loads all commands from a given script.
func Load(r io.Reader) error {
	return provider.Load(r)
}

// LoadDir loads all script commands from a given directory. Note that all
// scripts should have .sql extension.
func LoadDir(dir string) error {
	return provider.LoadDir(dir)
}

// Command returns a command for given name and parameters. The operation can
// panic if the command cannot be found.
func Command(name string, params ...script.Param) *script.Cmd {
	cmd, err := provider.Command(name, params...)

	if err != nil {
		panic(err)
	}

	return cmd
}

// SQL create a new command from raw query
func SQL(query string, params ...script.Param) *script.Cmd {
	return script.SQL(query, params...)
}
