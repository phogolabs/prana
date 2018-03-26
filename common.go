package gom

import (
	"database/sql"
	"io"

	"github.com/jmoiron/sqlx"
)

var provider *CmdProvider

type Param = interface{}
type Entity = interface{}
type Rows = sqlx.Rows
type Row = sqlx.Row
type Result = sql.Result
type Tx = sqlx.Tx

func init() {
	provider = &CmdProvider{
		Repository: make(map[string]string),
	}
}

func Load(r io.Reader) error {
	return provider.Load(r)
}

func Command(name string, params ...Param) *Cmd {
	cmd, err := provider.Command(name, params...)

	if err != nil {
		panic(err)
	}

	return cmd
}
