package sqlmigr

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/apex/log"
)

// Executor provides a group of operations that works with sqlmigrs.
type Executor struct {
	// Logger logs each execution step
	Logger log.Interface
	// Provider provides all sqlmigrs for the current project.
	Provider MigrationProvider
	// Runner runs or reverts sqlmigrs for the current project.
	Runner MigrationRunner
	// Generator generates a sqlmigr file.
	Generator MigrationGenerator
}

// Setup setups the current project for database sqlmigrs by creating
// sqlmigr directory and related database.
func (m *Executor) Setup() error {
	sqlmigr := &Migration{
		ID:          min.Format(format),
		Description: "setup",
		CreatedAt:   time.Now(),
	}

	if ok := m.Provider.Exists(sqlmigr); ok {
		return nil
	}

	up := &bytes.Buffer{}
	fmt.Fprintln(up, "CREATE TABLE IF NOT EXISTS migrations (")
	fmt.Fprintln(up, " id          TEXT      NOT NULL PRIMARY KEY,")
	fmt.Fprintln(up, " description TEXT      NOT NULL,")
	fmt.Fprintln(up, " created_at  TIMESTAMP NOT NULL")
	fmt.Fprintln(up, ");")

	down := bytes.NewBufferString("DROP TABLE IF EXISTS migrations;")

	content := &Content{
		UpCommand:   up,
		DownCommand: down,
	}

	if err := m.Generator.Write(sqlmigr, content); err != nil && !os.IsExist(err) {
		return err
	}

	if err := m.Runner.Run(sqlmigr); err != nil {
		return err
	}

	return m.Provider.Insert(sqlmigr)
}

// Create creates a sqlmigr script successfully if the project has already
// been setup, otherwise returns an error.
func (m *Executor) Create(name string) (*Migration, error) {
	name = strings.Replace(name, " ", "_", -1)

	timestamp := time.Now()

	sqlmigr := &Migration{
		ID:          timestamp.Format(format),
		Description: name,
		CreatedAt:   timestamp,
	}

	if err := m.Generator.Create(sqlmigr); err != nil {
		return nil, err
	}

	return sqlmigr, nil
}

// Run runs a pending sqlmigr for given count. If the count is negative number, it
// will execute all pending sqlmigrs.
func (m *Executor) Run(step int) (int, error) {
	run := 0
	sqlmigrs, err := m.Migrations()
	if err != nil {
		return run, err
	}

	m.logf("Running migration(s)")
	for _, sqlmigr := range sqlmigrs {
		if step == 0 {
			return run, nil
		}

		timestamp, err := time.Parse(format, sqlmigr.ID)
		if err != nil {
			return run, err
		}

		if !sqlmigr.CreatedAt.IsZero() || timestamp == min {
			continue
		}

		op := sqlmigr

		m.logf("Running migration '%s'", sqlmigr.Filename())

		if err := m.Runner.Run(&op); err != nil {
			return run, err
		}

		if err := m.Provider.Insert(&op); err != nil {
			return run, err
		}

		step = step - 1
		run = run + 1
	}

	m.logf("Run %d sqlmigr(s)", run)
	return run, nil
}

// RunAll runs all pending sqlmigrs.
func (m *Executor) RunAll() (int, error) {
	return m.Run(-1)
}

// Revert reverts an applied sqlmigr for given count. If the count is
// negative number, it will revert all applied sqlmigrs.
func (m *Executor) Revert(step int) (int, error) {
	reverted := 0
	sqlmigrs, err := m.Migrations()
	if err != nil {
		return reverted, err
	}

	m.logf("Reverting sqlmigr(s)")
	for i := len(sqlmigrs) - 1; i >= 0; i-- {
		sqlmigr := sqlmigrs[i]

		if step == 0 {
			return reverted, nil
		}

		if sqlmigr.CreatedAt.IsZero() {
			continue
		}

		timestamp, err := time.Parse(format, sqlmigr.ID)
		if err != nil || timestamp == min {
			return reverted, err
		}

		op := sqlmigr

		m.logf("Reverting sqlmigr '%s'", sqlmigr.Filename())
		if err := m.Runner.Revert(&op); err != nil {
			return reverted, err
		}

		if err := m.Provider.Delete(&op); err != nil {
			return reverted, err
		}

		step = step - 1
		reverted = reverted + 1
	}

	m.logf("Reverted %d sqlmigr(s)", reverted)
	return reverted, nil
}

// RevertAll reverts all applied sqlmigrs.
func (m *Executor) RevertAll() (int, error) {
	return m.Revert(-1)
}

// Migrations returns all sqlmigrs.
func (m *Executor) Migrations() ([]Migration, error) {
	return m.Provider.Migrations()
}

func (m *Executor) logf(text string, args ...interface{}) {
	if m.Logger != nil {
		m.Logger.Infof(text, args...)
	}
}
