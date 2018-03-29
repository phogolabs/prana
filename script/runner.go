package script

import "github.com/jmoiron/sqlx"

type Runner struct {
	Dir string
	DB  *sqlx.DB
}

func (r *Runner) Run(name string, args ...Param) (*Rows, error) {
	provider := &Provider{}

	if err := provider.LoadDir(r.Dir); err != nil {
		return nil, err
	}

	cmd, err := provider.Command(name, args...)
	if err != nil {
		return nil, err
	}

	query, params := cmd.Prepare()

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
