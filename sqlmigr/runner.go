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
	return r.exec("up", m)
}

// Revert reverts a given migration  item.
func (r *Runner) Revert(m *Migration) error {
	return r.exec("down", m)
}

func (r *Runner) exec(step string, m *Migration) error {
	statements, err := r.routine(step, m)
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
				Statement: query,
			}
		}
	}

	return tx.Commit()
}

func (r *Runner) routine(name string, m *Migration) ([]string, error) {
	statements := make(map[string]string, 2)

	for _, file := range m.Filenames() {
		routines, err := r.scan(file)
		if err != nil {
			return []string{}, err
		}

		for key, value := range routines {
			statements[key] = value
		}
	}

	routine, ok := statements[name]
	if !ok {
		return []string{}, fmt.Errorf("routine '%s' not found for migration '%v'", name, m)
	}

	commands := strings.FieldsFunc(routine, func(c rune) bool {
		return c == ';'
	})

	for index, cmd := range commands {
		commands[index] = strings.TrimSpace(cmd)
	}

	return commands, nil
}

func (r *Runner) scan(filename string) (map[string]string, error) {
	file, err := r.FileSystem.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	defer func() {
		if ioErr := file.Close(); err == nil {
			err = ioErr
		}
	}()

	scanner := &sqlexec.Scanner{}
	return scanner.Scan(file), nil
}
