package gom

import (
	"github.com/jmoiron/sqlx"
)

// Gateway is connected to a database and can executes SQL queries againts it.
type Gateway struct {
	db *sqlx.DB
}

// Open creates a new gateway connected to the provided source.
func Open(driver, source string) (*Gateway, error) {
	db, err := sqlx.Open(driver, source)
	if err != nil {
		return nil, err
	}

	return &Gateway{db: db}, nil
}

// Close closes the connection to underlying database.
func (g *Gateway) Close() error {
	return g.db.Close()
}

// Select executes a given query and maps the result to the provided slice of entities.
func (g *Gateway) Select(dest Entity, query Query) error {
	return selectMany(g.db, dest, query)
}

// Select executes a given query and maps a single result to the provided entity.
func (g *Gateway) SelectOne(dest Entity, query Query) error {
	return selectOne(g.db, dest, query)
}

// Query executes a given query and returns an instance of rows cursor.
func (g *Gateway) Query(query Query) (*Rows, error) {
	return queryRows(g.db, query)
}

// Query executes a given query and returns an instance of row.
func (g *Gateway) QueryRow(query Query) (*Row, error) {
	return queryRow(g.db, query)
}

// Exec executes a given query. It returns a result that provides information
// about the affected rows.
func (g *Gateway) Exec(query Query) (Result, error) {
	return exec(g.db, query)
}
