package migration

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

// Generator generates a new migration file for given directory.
type Generator struct {
	// FileSystem is the file system where all migrations are created.
	FileSystem FileSystem
}

// Create creates a new migration.
func (g *Generator) Create(m *Item) (string, error) {
	if err := g.Write(m, nil); err != nil {
		return "", err
	}

	return g.FileSystem.Join(m.Filename()), nil
}

// Write creates a new migration for given content.
func (g *Generator) Write(m *Item, content *Content) error {
	buffer := &bytes.Buffer{}

	fmt.Fprintln(buffer, "-- Auto-generated at", m.CreatedAt.Format(time.UnixDate))
	fmt.Fprintln(buffer, "-- Please do not change the name attributes")
	fmt.Fprintln(buffer)
	fmt.Fprintln(buffer, "-- name: up")
	fmt.Fprintln(buffer)

	if content != nil {
		if _, err := io.Copy(buffer, content.UpCommand); err != nil {
			return err
		}
	}

	fmt.Fprintln(buffer, "-- name: down")
	fmt.Fprintln(buffer)

	if content != nil {
		if _, err := io.Copy(buffer, content.DownCommand); err != nil {
			return err
		}
	}

	if err := g.FileSystem.WriteFile(m.Filename(), buffer.Bytes(), 0600); err != nil {
		return err
	}

	return nil
}
