package gom

import (
	"database/sql"
	"io"

	"github.com/jmoiron/sqlx"
	"github.com/svett/gom/script"
)

type Entity = interface{}
type Rows = sqlx.Rows
type Row = sqlx.Row
type Result = sql.Result
type Tx = sqlx.Tx

var provider *script.Provider

func init() {
	provider = &script.Provider{}
}

func Load(r io.Reader) error {
	return provider.Load(r)
}

func LoadDir(dir string) error {
	return provider.LoadDir(dir)
}

func Command(name string, params ...script.Param) *script.Cmd {
	cmd, err := provider.Command(name, params...)

	if err != nil {
		panic(err)
	}

	return cmd
}
