package script

import "github.com/jmoiron/sqlx"

var (
	format = "20060102150405"
)

type Rows = sqlx.Rows
type Param = interface{}

type Query interface {
	Prepare() (string, map[string]interface{})
}

type Gateway interface {
	Query(preparer Query) (*Rows, error)
	Close() error
}
