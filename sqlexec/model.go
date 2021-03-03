// Package sqlexec provides primitives and functions to work with raw SQL
// statements and pre-defined SQL Scripts.
package sqlexec

import (
	"io/fs"

	"github.com/jmoiron/sqlx"
)

var (
	format = "20060102150405"
)

// Param is a command parameter for given query.
type Param = interface{}

// Rows is a wrapper around sql.Rows which caches costly reflect operations
// during a looped StructScan.
type Rows = sqlx.Rows

// FileSystem provides with primitives to work with the underlying file system
type FileSystem = fs.FS

// WriteFileSystem represents a wriable file system
type WriteFileSystem interface {
	FileSystem

	// OpenFile opens a new file
	OpenFile(string, int, fs.FileMode) (fs.File, error)
}
