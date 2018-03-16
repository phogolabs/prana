package gom

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var (
	DateTimeFormat = "20060102150405"
	MinTime        = time.Date(1, time.January, 1970, 0, 0, 0, 0, time.UTC)
)

type Migrator struct {
	Dir     string
	Gateway *Gateway
}

func (m *Migrator) Setup() error {
	directories := []string{
		filepath.Join(m.Dir, "/database/migration"),
		filepath.Join(m.Dir, "/database/statement"),
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}

	buffer := &bytes.Buffer{}
	fmt.Fprintln(buffer, "-- Auto-generated at", time.Now().UTC().Format(time.UnixDate))
	fmt.Fprintln(buffer, "-- Please do not modify the contents")
	fmt.Fprintln(buffer)
	fmt.Fprintln(buffer, "-- name: up")
	fmt.Fprintln(buffer, "CREATE TABLE system.migrations (")
	fmt.Fprintln(buffer, " id TEXT NOT NULL PRIMARY KEY,")
	fmt.Fprintln(buffer, " updated_at TIMESTAMP DEFAULT now() NOT NULL,")
	fmt.Fprintln(buffer, " created_at TIMESTAMP DEFAULT now() NOT NULL,")
	fmt.Fprintln(buffer, ")")
	fmt.Fprintln(buffer)
	fmt.Fprintln(buffer, "-- name: down")
	fmt.Fprintln(buffer, "DROP TABLE IF EXISTS system.migrations")

	path := fmt.Sprintf("/database/migration/%s_setup.sql", MinTime.Format(DateTimeFormat))
	path = filepath.Join(m.Dir, path)

	_, err := os.Stat(path)
	if err == nil {
		return fmt.Errorf("The project has already been configured")
	}

	if err := ioutil.WriteFile(path, buffer.Bytes(), 0600); err != nil {
		return err
	}

	return nil
}

func (m *Migrator) Create(name string) (string, error) {
	if err := m.configured(); err != nil {
		return "", err
	}

	timestamp := time.Now().UTC()
	buffer := &bytes.Buffer{}

	fmt.Fprintln(buffer, "-- Auto-generated at", timestamp.Format(time.UnixDate))
	fmt.Fprintln(buffer)
	fmt.Fprintln(buffer, "-- name: up")
	fmt.Fprintln(buffer)
	fmt.Fprintln(buffer, "-- name: down")
	fmt.Fprintln(buffer)

	path := fmt.Sprintf("/database/migration/%s_%s.sql", timestamp.Format(DateTimeFormat), name)
	path = filepath.Join(m.Dir, path)

	if err := ioutil.WriteFile(path, buffer.Bytes(), 0600); err != nil {
		return "", err
	}

	return path, nil
}

func (m *Migrator) configured() error {
	path := fmt.Sprintf("/database/migration/%s_setup.sql", MinTime.Format(DateTimeFormat))
	path = filepath.Join(m.Dir, path)

	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("The project has not been configured")
	}

	return nil
}
