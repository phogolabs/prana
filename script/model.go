// Package script provides primitives and functions to work with raw SQL
// statements and pre-defined SQL Scripts.
package script

import (
	"io"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
)

var (
	format = "20060102150405"
)

//go:generate counterfeiter -fake-name ScriptFileSystem -o ../fake/ScriptFileSystem.go . FileSystem

// Rows is a wrapper around sql.Rows which caches costly reflect operations
// during a looped StructScan.
type Rows = sqlx.Rows

// Param is a command parameter for given query.
type Param = interface{}

// FileSystem provides with primitives to work with the underlying file system
type FileSystem interface {
	// Walk walks the file tree rooted at root, calling walkFn for each file or
	// directory in the tree, including root.
	Walk(dir string, fn filepath.WalkFunc) error
	// OpenFile is the generalized open call; most users will use Open
	OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error)
	// MkdirAll creates a directory named path
	MkdirAll(dir string, perm os.FileMode) error
}
