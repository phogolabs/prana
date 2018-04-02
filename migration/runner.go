package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/phogolabs/gom/script"
)

// Runner runs or reverts a given migration item.
type Runner struct {
	// Dir represents the project directory.
	Dir string
	// DB is a client to underlying database.
	DB *sqlx.DB
}

// Run runs a given migration item.
func (r *Runner) Run(m *Item) error {
	if err := r.exec("up", m); err != nil {
		return err
	}

	return nil
}

// Revert reverts a given migration item.
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
	path := filepath.Join(r.Dir, m.Filename())
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		if ioErr := file.Close(); err == nil {
			err = ioErr
		}
	}()

	scanner := &script.Scanner{}

	queries := scanner.Scan(file)
	statements, ok := queries[name]

	if !ok {
		return []string{}, fmt.Errorf("Command '%s' not found", name)
	}

	commands := strings.Split(statements, ";")
	return commands, nil
}
