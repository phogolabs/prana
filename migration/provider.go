package migration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/svett/gom"
)

type Provider struct {
	Dir     string
	Gateway *gom.Gateway
}

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
