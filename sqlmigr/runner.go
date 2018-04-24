package sqlmigr

import (
	"fmt"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/phogolabs/oak/sqlexec"
)

// Runner runs or reverts a given sqlmigr item.
type Runner struct {
	// FileSystem represents the project directory file system.
	FileSystem FileSystem
	// DB is a client to underlying database.
	DB *sqlx.DB
}

// Run runs a given sqlmigr item.
func (r *Runner) Run(m *Item) error {
	if err := r.exec("up", m); err != nil {
		return err
	}

	return nil
}

// Revert reverts a given sqlmigr item.
func (r *Runner) Revert(m *Item) error {
	if err := r.exec("down", m); err != nil {
		return err
	}
	return nil
}

func (r *Runner) exec(step string, m *Item) error {
	statements, err := r.command(step, m)
	if err != nil {
		return err
	}

	for _, query := range statements {
		if _, err := r.DB.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) command(name string, m *Item) ([]string, error) {
	file, err := r.FileSystem.OpenFile(m.Filename(), os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	defer func() {
		if ioErr := file.Close(); err == nil {
			err = ioErr
		}
	}()

	scanner := &sqlexec.Scanner{}

	queries := scanner.Scan(file)
	statements, ok := queries[name]

	if !ok {
		return []string{}, fmt.Errorf("Command '%s' not found for sqlmigr '%s'", name, m.Filename())
	}

	commands := strings.Split(statements, ";")
	return commands, nil
}
