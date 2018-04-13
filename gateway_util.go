package oak

import "github.com/jmoiron/sqlx"

func selectMany(preparer Preparer, dest Entity, query Query) error {
	stmt, args, err := prepareQuery(preparer, query)
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

func selectOne(preparer Preparer, dest Entity, query Query) error {
	stmt, args, err := prepareQuery(preparer, query)
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

func queryRows(preparer Preparer, query Query) (*Rows, error) {
	stmt, args, err := prepareQuery(preparer, query)
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

func queryRow(preparer Preparer, query Query) (*Row, error) {
	stmt, args, err := prepareQuery(preparer, query)
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

func exec(preparer Preparer, query Query) (Result, error) {
	stmt, args, err := prepareQuery(preparer, query)
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

func prepareQuery(preparer Preparer, query Query) (*sqlx.NamedStmt, map[string]interface{}, error) {
	body, args := query.Prepare()

	stmt, err := preparer.PrepareNamed(body)
	if err != nil {
		return nil, nil, err
	}

	return stmt, args, nil
}
