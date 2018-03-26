package migration

import (
	"bytes"
	"fmt"
	"time"
)

type Executor struct {
	Provider   ItemProvider
	Runner     ItemRunner
	Generator  FileGenerator
	OnRunFn    RunFn
	OnRevertFn RevertFn
}

func (m *Executor) Setup() error {
	migration := &Item{
		Id:          min.Format(format),
		Description: "setup",
		CreatedAt:   time.Now(),
	}

	up := &bytes.Buffer{}
	fmt.Fprintln(up, "CREATE TABLE migrations (")
	fmt.Fprintln(up, " id          TEXT      NOT NULL PRIMARY KEY,")
	fmt.Fprintln(up, " description TEXT      NOT NULL,")
	fmt.Fprintln(up, " created_at  TIMESTAMP NOT NULL")
	fmt.Fprintln(up, ");")

	down := bytes.NewBufferString("DROP TABLE IF EXISTS migrations;")

	content := &Content{
		UpCommand:   up,
		DownCommand: down,
	}

	if err := m.Generator.Write(migration, content); err != nil {
		return err
	}

	return m.Runner.Run(migration)
}

func (m *Executor) Create(name string) (string, error) {
	timestamp := time.Now()

	migration := &Item{
		Id:          timestamp.Format(format),
		Description: name,
		CreatedAt:   timestamp,
	}

	return m.Generator.Create(migration)
}

func (m *Executor) Run(step int) error {
	migrations, err := m.Migrations()
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if step == 0 {
			return nil
		}

		timestamp, err := time.Parse(format, migration.Id)
		if err != nil {
			return err
		}

		if !migration.CreatedAt.IsZero() || timestamp == min {
			continue
		}

		op := migration

		if m.OnRunFn != nil {
			m.OnRunFn(&op)
		}

		if err := m.Runner.Run(&op); err != nil {
			return err
		}

		step = step - 1
	}

	return nil
}

func (m *Executor) Revert(step int) error {
	migrations, err := m.Migrations()
	if err != nil {
		return err
	}

	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]

		if step == 0 {
			return nil
		}

		if migration.CreatedAt.IsZero() {
			continue
		}

		timestamp, err := time.Parse(format, migration.Id)
		if err != nil || timestamp == min {
			return err
		}

		op := migration

		if m.OnRevertFn != nil {
			m.OnRevertFn(&op)
		}

		if err := m.Runner.Revert(&op); err != nil {
			return err
		}

		step = step - 1
	}

	return nil
}

func (m *Executor) Migrations() ([]Item, error) {
	return m.Provider.Migrations()
}
