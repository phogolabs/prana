package sqlmigr

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"
)

var _ MigrationGenerator = &Generator{}

// Generator generates a new sqlmigr file for given directory.
type Generator struct {
	// FileSystem is the file system where all sqlmigrs are created.
	FileSystem WriteFileSystem
}

// Create creates a new sqlmigr.
func (g *Generator) Create(m *Migration) error {
	return g.Write(m, nil)
}

// Write creates a new sqlmigr for given content.
func (g *Generator) Write(m *Migration, content *Content) error {
	buffer := &bytes.Buffer{}

	fmt.Fprintln(buffer, "-- Auto-generated at", m.CreatedAt.Format(time.RFC1123))
	fmt.Fprintln(buffer, "-- Please do not change the name attributes")
	fmt.Fprintln(buffer)
	fmt.Fprintln(buffer, "-- name: up")

	if content != nil {
		if _, err := io.Copy(buffer, content.UpCommand); err != nil {
			return err
		}
	} else {
		fmt.Fprintln(buffer)
	}

	fmt.Fprintln(buffer, "-- name: down")

	if content != nil {
		if _, err := io.Copy(buffer, content.DownCommand); err != nil {
			return err
		}
	} else {
		fmt.Fprintln(buffer)
	}

	for _, filename := range m.Filenames() {
		if err := g.write(filename, buffer.Bytes(), 0600); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) write(filename string, data []byte, perm os.FileMode) error {
	f, err := g.FileSystem.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}

	if writer, ok := f.(io.Writer); ok {
		n, err := writer.Write(data)
		if err == nil && n < len(data) {
			err = io.ErrShortWrite
		}
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
