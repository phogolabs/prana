package sqlexec

import "github.com/jmoiron/sqlx"

// Runner runs a SQL statement for given command name and parameters.
type Runner struct {
	// FileSystem represents the project directory file system.
	FileSystem FileSystem
	// DB is a client to underlying database.
	DB *sqlx.DB
}

// Run runs a given command with provided parameters.
func (r *Runner) Run(name string, args ...Param) (*Rows, error) {
	provider := &Provider{}

	if err := provider.ReadDir(r.FileSystem); err != nil {
		return nil, err
	}

	cmd, err := provider.Query(name, args...)
	if err != nil {
		return nil, err
	}

	query, params := cmd.NamedQuery()

	stmt, err := r.DB.PrepareNamed(query)
	if err != nil {
		return nil, err
	}

	defer func() {
		if stmtErr := stmt.Close(); err == nil {
			err = stmtErr
		}
	}()

	var rows *Rows
	rows, err = stmt.Queryx(params)
	return rows, err
}
