package sqlmigr

import (
	"fmt"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/phogolabs/prana/sqlexec"
)

var _ MigrationRunner = &Runner{}

// Runner runs or reverts a given migration  item.
type Runner struct {
	// FileSystem represents the project directory file system.
	FileSystem FileSystem
	// DB is a client to underlying database.
	DB *sqlx.DB
}

// Run runs a given migration  item.
func (r *Runner) Run(m *Migration) error {
	if err := r.exec("up", m); err != nil {
		return err
	}

	return nil
}

// Revert reverts a given migration  item.
func (r *Runner) Revert(m *Migration) error {
	if err := r.exec("down", m); err != nil {
		return err
	}
	return nil
}

func (r *Runner) exec(step string, m *Migration) error {
	statements, err := r.command(step, m)
	if err != nil {
		return err
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	for _, query := range statements {
		if _, err := tx.Exec(query); err != nil {
			tx.Rollback()

			return &RunnerError{
				Err:       err,
				Migration: m.Filename(),
				Statement: query,
			}
		}
	}

	return tx.Commit()
}

func (r *Runner) command(name string, m *Migration) ([]string, error) {
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
		return []string{}, fmt.Errorf("Routine '%s' not found for migration '%s'", name, m.Filename())
	}

	commands := strings.FieldsFunc(statements, func(c rune) bool {
		return c == ';'
	})

	for index, cmd := range commands {
		commands[index] = strings.TrimSpace(cmd)
	}

	return commands, nil
}
