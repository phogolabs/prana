package sqlexec

import (
	"fmt"
	"io"

	"github.com/jmoiron/sqlx"
	"github.com/olekukonko/tablewriter"
)

// Runner runs a SQL statement for given command name and parameters.
type Runner struct {
	// FileSystem represents the project directory file system.
	FileSystem FileSystem
	// DB is a client to underlying database.
	DB *sqlx.DB
}

// Run runs a given command with provided parameters.
func (r *Runner) Run(name string, args ...Param) (*Rows, error) {
	provider := &Provider{
		DriverName: r.DB.DriverName(),
	}

	if err := provider.ReadDir(r.FileSystem); err != nil {
		return nil, err
	}

	query, err := provider.Query(name)
	if err != nil {
		return nil, err
	}

	stmt, err := r.DB.Preparex(query)
	if err != nil {
		return nil, err
	}

	defer func() {
		if stmtErr := stmt.Close(); err == nil {
			err = stmtErr
		}
	}()

	return stmt.Queryx(args...)
}

// Print prints the rows
func (r *Runner) Print(writer io.Writer, rows *sqlx.Rows) error {
	table := tablewriter.NewWriter(writer)

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	table.SetHeader(columns)

	for rows.Next() {
		record, err := rows.SliceScan()
		if err != nil {
			return err
		}

		row := []string{}

		for _, column := range record {
			if data, ok := column.([]byte); ok {
				column = string(data)
			}
			row = append(row, fmt.Sprintf("%v", column))
		}

		table.Append(row)
	}

	table.Render()
	return nil
}
