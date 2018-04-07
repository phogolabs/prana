package gom

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Query represents an SQL Query that can be executed by Gateway.
type Query interface {
	// Prepare prepares the query for execution. It returns the actual query and
	// a maps of its arguments.
	Prepare() (string, map[string]interface{})
}

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
func (g *Gateway) Select(dest Entity, preparer Query) error {
	stmt, args, err := g.prepare(preparer)
	if err != nil {
		return err
	}

	defer func() {
		if stmtErr := stmt.Close(); err == nil {
			err = stmtErr
		}
	}()

	err = stmt.Select(dest, args)
	return err
}

// Select executes a given query and maps a single result to the provided entity.
func (g *Gateway) SelectOne(dest Entity, preparer Query) error {
	stmt, args, err := g.prepare(preparer)
	if err != nil {
		return err
	}

	defer func() {
		if stmtErr := stmt.Close(); err == nil {
			err = stmtErr
		}
	}()

	err = stmt.Get(dest, args)
	return err
}

// Query executes a given query and returns an instance of rows cursor.
func (g *Gateway) Query(preparer Query) (*Rows, error) {
	stmt, args, err := g.prepare(preparer)
	if err != nil {
		return nil, err
	}

	defer func() {
		if stmtErr := stmt.Close(); err == nil {
			err = stmtErr
		}
	}()

	var rows *Rows
	rows, err = stmt.Queryx(args)
	return rows, err
}

// Query executes a given query and returns an instance of row.
func (g *Gateway) QueryRow(preparer Query) (*Row, error) {
	stmt, args, err := g.prepare(preparer)
	if err != nil {
		return nil, err
	}

	defer func() {
		if stmtErr := stmt.Close(); err == nil {
			err = stmtErr
		}
	}()

	return stmt.QueryRowx(args), nil
}

// Exec executes a given query. It returns a result that provides information
// about the affected rows.
func (g *Gateway) Exec(preparer Query) (Result, error) {
	stmt, args, err := g.prepare(preparer)
	if err != nil {
		return nil, err
	}

	defer func() {
		if stmtErr := stmt.Close(); err == nil {
			err = stmtErr
		}
	}()

	var result Result
	result, err = stmt.Exec(args)
	return result, err
}

func (g *Gateway) prepare(preparer Query) (*sqlx.NamedStmt, map[string]interface{}, error) {
	query, args := preparer.Prepare()

	stmt, err := g.db.PrepareNamed(query)
	if err != nil {
		return nil, nil, err
	}

	return stmt, args, nil
}
