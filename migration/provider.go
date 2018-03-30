package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/phogolabs/gom"
)

// Provider provides all migration for given project.
type Provider struct {
	// Dir represents the project directory.
	Dir string
	// Gateway is a client to underlying database.
	Gateway *gom.Gateway
}

// Migrations returns the project migrations.
func (m *Provider) Migrations() ([]Item, error) {
	migrations := []Item{}

	err := filepath.Walk(m.Dir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return fmt.Errorf("Directory '%s' does not exist", m.Dir)
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
	query := gom.Select("id", "description", "created_at").
		From("migrations").
		OrderBy(gom.Order("id", gom.Asc))

	if err := m.Gateway.Select(&applied, query); err != nil {
		return []Item{}, err
	}

	for index, migration := range applied {
		m := migrations[index]

		if m.Id != migration.Id {
			err = fmt.Errorf("Mismatched migration id. Expected: '%s' but has '%s'", migration.Id, m.Id)
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
	item.CreatedAt = time.Now()

	query := gom.Insert("migrations").
		Set(
			gom.Pair("id", item.Id),
			gom.Pair("description", item.Description),
			gom.Pair("created_at", item.CreatedAt),
		)

	if _, err := m.Gateway.Exec(query); err != nil {
		return err
	}

	return nil
}

// Delete deletes applied migration item from migrations table.
func (m *Provider) Delete(item *Item) error {
	query := gom.Delete("migrations").Where(gom.Condition("id").Equal(item.Id))

	if _, err := m.Gateway.Exec(query); err != nil {
		return err
	}

	return nil
}
