// Package sqlexec provides primitives and functions to work with raw SQL
// statements and pre-defined SQL Scripts.
package sqlexec

import (
	"github.com/jmoiron/sqlx"
	"github.com/phogolabs/parcello"
)

var (
	format = "20060102150405"
)

// Rows is a wrapper around sql.Rows which caches costly reflect operations
// during a looped StructScan.
type Rows = sqlx.Rows

// Param is a command parameter for given query.
type Param = interface{}

// P is a shortcut to a map. It facilitates passing named params to a named
// commands and queries
type P = map[string]Param

// FileSystem provides with primitives to work with the underlying file system
type FileSystem = parcello.FileSystem
