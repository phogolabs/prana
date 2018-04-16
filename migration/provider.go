package migration

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
)

// Provider provides all migration for given project.
type Provider struct {
	// FileSystem represents the project directory file system.
	FileSystem FileSystem
	// DB is a client to underlying database.
	DB *sqlx.DB
}

// Migrations returns the project migrations.
func (m *Provider) Migrations() ([]Item, error) {
	migrations := []Item{}

	err := m.FileSystem.Walk("/", func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return fmt.Errorf("Directory '%s' does not exist", path)
		}

		if info.IsDir() {
			return nil
		}

		matched, err := filepath.Match("*.sql", info.Name())
		if err != nil || !matched {
			return err
		}

		migration, err := Parse(path)
		if err != nil {
			return err
		}

		migrations = append(migrations, *migration)
		return nil
	})

	if err != nil {
		return []Item{}, err
	}

	applied := []Item{}

	query := &bytes.Buffer{}
	query.WriteString("SELECT id, description, created_at ")
	query.WriteString("FROM migrations ")
	query.WriteString("ORDER BY id ASC")

	if err := m.DB.Select(&applied, query.String()); err != nil {
		return []Item{}, err
	}

	for index, migration := range applied {
		m := migrations[index]

		if m.ID != migration.ID {
			err = fmt.Errorf("Mismatched migration id. Expected: '%s' but has '%s'", migration.ID, m.ID)
			return []Item{}, err
		}

		if m.Description != migration.Description {
			err = fmt.Errorf("Mismatched migration description. Expected: '%s' but has '%s'", migration.Description, m.Description)
			return []Item{}, err
		}

		migrations[index] = migration
	}

	return migrations, err
}

// Insert inserts exectued migration item in the migrations table.
func (m *Provider) Insert(item *Item) error {
	rows, err := m.DB.Query("SELECT id FROM migrations WHERE id = ?", item.ID)

	if err != nil {
		return err
	}

	defer rows.Close()

	if rows.Next() {
		return nil
	}

	item.CreatedAt = time.Now()

	query := &bytes.Buffer{}
	query.WriteString("INSERT INTO migrations(id, description, created_at) ")
	query.WriteString("VALUES (?, ?, ?)")

	if _, err := m.DB.Exec(query.String(), item.ID, item.Description, item.CreatedAt); err != nil {
		return err
	}

	return nil
}

// Delete deletes applied migration item from migrations table.
func (m *Provider) Delete(item *Item) error {
	query := &bytes.Buffer{}
	query.WriteString("DELETE FROM migrations ")
	query.WriteString("WHERE id = ?")

	if _, err := m.DB.Exec(query.String(), item.ID); err != nil {
		return err
	}

	return nil
}
