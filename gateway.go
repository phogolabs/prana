package oak

import (
	"github.com/jmoiron/sqlx"
)

// Gateway is connected to a database and can executes SQL queries against it.
type Gateway struct {
	db *sqlx.DB
}

// OpenURL creates a new gateway connecto to the provided URL.
func OpenURL(url string) (*Gateway, error) {
	driver, source, err := ParseURL(url)
	if err != nil {
		return nil, err
	}

	return Open(driver, source)
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

// DriverName returns the driverName passed to the Open function for this DB.
func (g *Gateway) DriverName() string {
	return g.db.DriverName()
}

// Begin begins a transaction and returns an *Tx
func (g *Gateway) Begin() (*Tx, error) {
	tx, err := g.db.Beginx()
	if err != nil {
		return nil, err
	}

	return &Tx{tx: tx}, nil
}

// Select executes a given query and maps the result to the provided slice of entities.
func (g *Gateway) Select(dest Entity, query Query) error {
	return selectMany(g.db, dest, query)
}

// SelectOne executes a given query and maps a single result to the provided entity.
func (g *Gateway) SelectOne(dest Entity, query Query) error {
	return selectOne(g.db, dest, query)
}

// Query executes a given query and returns an instance of rows cursor.
func (g *Gateway) Query(query Query) (*Rows, error) {
	return queryRows(g.db, query)
}

// QueryRow executes a given query and returns an instance of row.
func (g *Gateway) QueryRow(query Query) (*Row, error) {
	return queryRow(g.db, query)
}

// Exec executes a given query. It returns a result that provides information
// about the affected rows.
func (g *Gateway) Exec(query Query) (Result, error) {
	return exec(g.db, query)
}

// Tx is an sqlx wrapper around sqlx.Tx with extra functionality
type Tx struct {
	tx *sqlx.Tx
}

// Commit commits the transaction.
func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

// Rollback aborts the transaction.
func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

// Select executes a given query and maps the result to the provided slice of entities.
func (tx *Tx) Select(dest Entity, query Query) error {
	return selectMany(tx.tx, dest, query)
}

// SelectOne executes a given query and maps a single result to the provided entity.
func (tx *Tx) SelectOne(dest Entity, query Query) error {
	return selectOne(tx.tx, dest, query)
}

// Query executes a given query and returns an instance of rows cursor.
func (tx *Tx) Query(query Query) (*Rows, error) {
	return queryRows(tx.tx, query)
}

// QueryRow executes a given query and returns an instance of row.
func (tx *Tx) QueryRow(query Query) (*Row, error) {
	return queryRow(tx.tx, query)
}

// Exec executes a given query. It returns a result that provides information
// about the affected rows.
func (tx *Tx) Exec(query Query) (Result, error) {
	return exec(tx.tx, query)
}
