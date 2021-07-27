package sqlmigr

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
)

var _ MigrationProvider = &Provider{}

// Provider provides all migration for given project.
type Provider struct {
	// FileSystem represents the project directory file system.
	FileSystem FileSystem
	// DB is a client to underlying database.
	DB *sqlx.DB
}

// Migrations returns the project migrations.
func (m *Provider) Migrations() ([]*Migration, error) {
	local, err := m.files()
	if err != nil {
		return local, err
	}

	remote, err := m.query()
	if err != nil {
		return remote, err
	}

	return m.merge(remote, local)
}

func (m *Provider) files() ([]*Migration, error) {
	local := []*Migration{}

	err := fs.WalkDir(m.FileSystem, ".", func(path string, info os.DirEntry, xerr error) error {
		if ferr := m.filter(info); ferr != nil {
			if ferr.Error() == "skip" {
				ferr = nil
			}

			return ferr
		}

		migration, err := Parse(path)
		if err != nil {
			return err
		}

		if !m.supported(migration.Drivers) {
			return nil
		}

		if index := len(local) - 1; index >= 0 {
			if prev := local[index]; migration.Equal(prev) {
				prev.Drivers = append(prev.Drivers, migration.Drivers...)
				local[index] = prev
				return nil
			}
		}

		local = append(local, migration)
		return nil
	})

	if err != nil {
		return []*Migration{}, err
	}

	return local, nil
}

func (m *Provider) filter(info fs.DirEntry) error {
	skip := fmt.Errorf("skip")

	if info == nil {
		return os.ErrNotExist
	}

	if info.IsDir() {
		return skip
	}

	matched, _ := filepath.Match("*.sql", info.Name())

	if !matched {
		return skip
	}

	return nil
}

func (m *Provider) supported(drivers []string) bool {
	for _, driver := range drivers {
		if driver == every || driver == m.DB.DriverName() {
			return true
		}
	}

	return false
}

func (m *Provider) query() ([]*Migration, error) {
	query := &bytes.Buffer{}
	query.WriteString("SELECT id, description, created_at ")
	query.WriteString("FROM " + m.table() + " ")
	query.WriteString("ORDER BY id ASC")

	remote := []*Migration{}

	if err := m.DB.Select(&remote, query.String()); err != nil && !IsNotExist(err) {
		return []*Migration{}, err
	}

	return remote, nil
}

// Insert inserts executed sqlmigr item in the sqlmigrs table.
func (m *Provider) Insert(item *Migration) error {
	item.CreatedAt = time.Now()

	builder := &bytes.Buffer{}
	builder.WriteString("INSERT INTO " + m.table() + "(id, description, created_at) ")
	builder.WriteString("VALUES (?, ?, ?)")

	query := m.DB.Rebind(builder.String())
	if _, err := m.DB.Exec(query, item.ID, item.Description, item.CreatedAt); err != nil {
		return err
	}

	return nil
}

// Delete deletes applied sqlmigr item from sqlmigrs table.
func (m *Provider) Delete(item *Migration) error {
	builder := &bytes.Buffer{}
	builder.WriteString("DELETE FROM " + m.table() + " ")
	builder.WriteString("WHERE id = ?")

	query := m.DB.Rebind(builder.String())
	if _, err := m.DB.Exec(query, item.ID); err != nil {
		return err
	}

	return nil
}

// Exists returns true if the sqlmigr exists
func (m *Provider) Exists(item *Migration) bool {
	count := 0

	if err := m.DB.Get(&count, "SELECT count(id) FROM "+m.table()+" WHERE id = ?", item.ID); err != nil {
		return false
	}

	return count == 1
}

func (m *Provider) merge(remote, local []*Migration) ([]*Migration, error) {
	result := local

	for index, r := range remote {
		l := local[index]

		if r.ID != l.ID {
			return []*Migration{}, fmt.Errorf("mismatched migration id. Expected: '%s' but has '%s'", r.ID, l.ID)
		}

		if r.Description != l.Description {
			return []*Migration{}, fmt.Errorf("mismatched migration description. Expected: '%s' but has '%s'", r.Description, l.Description)
		}

		// Merge creation time
		l.CreatedAt = r.CreatedAt
		result[index] = l
	}

	return result, nil
}

func (m *Provider) table() string {
	for _, path := range setup.Filenames() {
		file, err := m.FileSystem.Open(path)
		if err != nil {
			continue
		}
		// close the file
		defer file.Close()

		if data, err := io.ReadAll(file); err == nil {
			if match := migrationRgxp.FindSubmatch(data); len(match) == 2 {
				return string(match[1])
			}
		}
	}

	return "migrations"
}
