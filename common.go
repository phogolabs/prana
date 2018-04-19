// Package oak provides a wrapper to work with loukoum built queries as well
// maitaining database version by creating, executing and reverting SQL
// migrations.
//
// The package allows executing embedded SQL statements from script for a given
// name.
package oak

import (
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/phogolabs/oak"
	"github.com/phogolabs/oak/migration"
	"github.com/phogolabs/oak/script"
	"github.com/phogolabs/parcel"
)

// Dir implements FileSystem using the native file system restricted to a
// specific directory tree.
type Dir = parcel.Dir

// Query represents an SQL Query that can be executed by Gateway.
type Query interface {
	// Prepare prepares the query for execution. It returns the actual query and
	// a maps of its arguments.
	Prepare() (string, map[string]interface{})
}

// Preparer prepares query for execution
type Preparer interface {
	// PrepareNamed returns a prepared named statement
	PrepareNamed(query string) (*NamedStmt, error)
}

// NamedStmt is a prepared statement that executes named queries.  Prepare it
// how you would execute a NamedQuery, but pass in a struct or map when executing.
type NamedStmt = sqlx.NamedStmt

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

var provider *script.Provider

func init() {
	provider = &script.Provider{}
}

// Setup setups the oak environment for us
func Setup(gateway *Gateway, manager *parcel.Manager) error {
	if err := oak.LoadSQLCommandsFrom(parcel.Root("script")); err != nil {
		return err
	}

	if err := oak.Migrate(gateway, parcel.Root("migration")); err != nil {
		return err
	}

	return nil
}

// Migrate runs all pending migration
func Migrate(gateway *Gateway, fileSystem parcel.FileSystem) error {
	return migration.RunAll(gateway.db, fileSystem)
}

// LoadSQLCommandsFromReader loads all commands from a given reader.
func LoadSQLCommandsFromReader(r io.Reader) error {
	_, err := provider.ReadFrom(r)
	return err
}

// LoadSQLCommandsFrom loads all script commands from a given directory. Note that all
// scripts should have .sql extension.
func LoadSQLCommandsFrom(fileSystem parcel.FileSystem) error {
	return provider.ReadDir(fileSystem)
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

// ParseURL parses a URL and returns the database driver and connection string to the database
func ParseURL(conn string) (string, string, error) {
	uri, err := url.Parse(conn)
	if err != nil {
		return "", "", err
	}

	driver := strings.ToLower(uri.Scheme)

	switch driver {
	case "mysql", "sqlite3":
		source := strings.Replace(conn, fmt.Sprintf("%s://", driver), "", -1)
		return driver, source, nil
	default:
		return driver, conn, nil
	}
}
