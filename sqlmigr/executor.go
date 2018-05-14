package sqlmigr

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/go-openapi/inflect"
)

// Executor provides a group of operations that works with migrations.
type Executor struct {
	// Logger logs each execution step
	Logger log.Interface
	// Provider provides all migrations for the current project.
	Provider MigrationProvider
	// Runner runs or reverts migrations for the current project.
	Runner MigrationRunner
	// Generator generates a migration file.
	Generator MigrationGenerator
}

// Setup setups the current project for database migrations by creating
// migration directory and related database.
func (m *Executor) Setup() error {
	migration := &Migration{
		ID:          min.Format(format),
		Description: "setup",
		Drivers:     []string{every},
		CreatedAt:   time.Now(),
	}

	up := &bytes.Buffer{}
	fmt.Fprintln(up, "CREATE TABLE IF NOT EXISTS migrations (")
	fmt.Fprintln(up, " id          VARCHAR(15) NOT NULL PRIMARY KEY,")
	fmt.Fprintln(up, " description TEXT        NOT NULL,")
	fmt.Fprintln(up, " created_at  TIMESTAMP   NOT NULL")
	fmt.Fprintln(up, ");")
	fmt.Fprintln(up)

	down := bytes.NewBufferString("DROP TABLE IF EXISTS migrations;")
	fmt.Fprintln(down)

	content := &Content{
		UpCommand:   up,
		DownCommand: down,
	}

	return m.Generator.Write(migration, content)
}

// Create creates a migration script successfully if the project has already
// been setup, otherwise returns an error.
func (m *Executor) Create(name string) (*Migration, error) {
	now := time.Now()
	id := now.Format(format)
	name = inflect.Underscore(strings.ToLower(name))
	name = fmt.Sprintf("%s_%s.sql", id, name)

	migration, _ := Parse(name)

	if migration != nil {
		migration.CreatedAt = now
	}

	if err := m.Generator.Create(migration); err != nil {
		return nil, err
	}

	return migration, nil
}

// Run runs a pending migration for given count. If the count is negative number, it
// will execute all pending migrations.
func (m *Executor) Run(step int) (int, error) {
	run := 0
	migrations, err := m.Migrations()
	if err != nil {
		return run, err
	}

	for index, migration := range migrations {
		if step == 0 {
			return run, nil
		}

		if !migration.CreatedAt.IsZero() {
			continue
		}

		m.logf("Running migration '%v'", migration)

		if err := m.Runner.Run(migrations[index]); err != nil {
			return run, err
		}

		if err := m.Provider.Insert(migrations[index]); err != nil {
			return run, err
		}

		step = step - 1
		run = run + 1
	}

	return run, nil
}

// RunAll runs all pending migrations.
func (m *Executor) RunAll() (int, error) {
	return m.Run(-1)
}

// Revert reverts an applied migration for given count. If the count is
// negative number, it will revert all applied migrations.
func (m *Executor) Revert(step int) (int, error) {
	reverted := 0
	migrations, err := m.Migrations()

	if err != nil {
		return reverted, err
	}

	for index := len(migrations) - 1; index >= 0; index-- {
		migration := migrations[index]

		if step == 0 {
			return reverted, nil
		}

		if migration.CreatedAt.IsZero() {
			continue
		}

		m.logf("Reverting migration '%v'", migration)

		if err := m.Runner.Revert(migrations[index]); err != nil {
			return reverted, err
		}

		if err := m.Provider.Delete(migrations[index]); err != nil {
			if IsNotExist(err) {
				err = nil
			}
			return reverted, err
		}

		step = step - 1
		reverted = reverted + 1
	}

	return reverted, nil
}

// RevertAll reverts all applied migrations.
func (m *Executor) RevertAll() (int, error) {
	return m.Revert(-1)
}

// Migrations returns all migrations.
func (m *Executor) Migrations() ([]*Migration, error) {
	return m.Provider.Migrations()
}

func (m *Executor) logf(text string, args ...interface{}) {
	if m.Logger != nil {
		m.Logger.Infof(text, args...)
	}
}
