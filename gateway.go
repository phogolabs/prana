package gom

import (
	"github.com/jmoiron/sqlx"
)

type Query interface {
	Prepare() (string, map[string]interface{})
}

type Gateway struct {
	db *sqlx.DB
}

func Open(driver, source string) (*Gateway, error) {
	db, err := sqlx.Open(driver, source)
	if err != nil {
		return nil, err
	}

	return &Gateway{db: db}, nil
}

func (g *Gateway) Close() error {
	return g.db.Close()
}

func (g *Gateway) DB() *sqlx.DB {
	return g.db
}

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

func (g *Gateway) SelectRow(dest Entity, preparer Query) error {
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
