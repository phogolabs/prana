package sqlmigr

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
)

// Provider provides all sqlmigr for given project.
type Provider struct {
	// FileSystem represents the project directory file system.
	FileSystem FileSystem
	// DB is a client to underlying database.
	DB *sqlx.DB
}

// Migrations returns the project sqlmigrs.
func (m *Provider) Migrations() ([]Item, error) {
	sqlmigrs := []Item{}

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

		sqlmigr, err := Parse(path)
		if err != nil {
			return err
		}

		sqlmigrs = append(sqlmigrs, *sqlmigr)
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

	for index, sqlmigr := range applied {
		m := sqlmigrs[index]

		if m.ID != sqlmigr.ID {
			err = fmt.Errorf("Mismatched sqlmigr id. Expected: '%s' but has '%s'", sqlmigr.ID, m.ID)
			return []Item{}, err
		}

		if m.Description != sqlmigr.Description {
			err = fmt.Errorf("Mismatched sqlmigr description. Expected: '%s' but has '%s'", sqlmigr.Description, m.Description)
			return []Item{}, err
		}

		sqlmigrs[index] = sqlmigr
	}

	return sqlmigrs, err
}

// Insert inserts executed sqlmigr item in the sqlmigrs table.
func (m *Provider) Insert(item *Item) error {
	item.CreatedAt = time.Now()

	query := &bytes.Buffer{}
	query.WriteString("INSERT INTO migrations(id, description, created_at) ")
	query.WriteString("VALUES (?, ?, ?)")

	if _, err := m.DB.Exec(query.String(), item.ID, item.Description, item.CreatedAt); err != nil {
		return err
	}

	return nil
}

// Delete deletes applied sqlmigr item from sqlmigrs table.
func (m *Provider) Delete(item *Item) error {
	query := &bytes.Buffer{}
	query.WriteString("DELETE FROM migrations ")
	query.WriteString("WHERE id = ?")

	if _, err := m.DB.Exec(query.String(), item.ID); err != nil {
		return err
	}

	return nil
}

// Exists returns true if the sqlmigr exists
func (m *Provider) Exists(item *Item) bool {
	count := 0

	if err := m.DB.Get(&count, "SELECT count(id) FROM migrations WHERE id = ?", item.ID); err != nil {
		return false
	}

	return count == 1
}
