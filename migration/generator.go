package migration

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type Generator struct {
	Dir string
}

func (g *Generator) Create(m *Item) (string, error) {
	if err := g.Write(m, nil); err != nil {
		return "", err
	}

	path := filepath.Join(g.Dir, m.Filename())
	return path, nil
}

func (g *Generator) Write(m *Item, content *Content) error {
	if err := os.MkdirAll(g.Dir, 0700); err != nil {
		return err
	}

	path := filepath.Join(g.Dir, m.Filename())

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("Migration '%s' already exists", path)
	}

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

	if err := ioutil.WriteFile(path, buffer.Bytes(), 0600); err != nil {
		return err
	}

	return nil
}
