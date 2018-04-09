package script

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Generator generates a new command.
type Generator struct {
	// Dir where the generates script will be stored
	Dir string
	// FileSystem represents the project directory file system.
	FileSystem FileSystem
}

// Create crates a new file and command for given file name and command name.
func (g *Generator) Create(container, name string) (string, error) {
	provider := &Provider{}

	if err := provider.ReadDir(g.Dir, g.FileSystem); err != nil {
		return "", err
	}

	if _, err := provider.Command(name); err == nil {
		return "", fmt.Errorf("Command '%s' already exists", name)
	}

	if container == "" {
		container = time.Now().Format(format)
	}

	path := filepath.Join(g.Dir, fmt.Sprintf("%s.sql", container))

	file, err := g.FileSystem.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return "", err
	}

	defer func() {
		if ioErr := file.Close(); err != nil {
			path = ""
			err = ioErr
		}
	}()

	fmt.Fprintln(file, "-- Auto-generated at", time.Now().Format(time.UnixDate))
	fmt.Fprintf(file, "-- name: %s", name)
	fmt.Fprintln(file)
	fmt.Fprintln(file)

	return path, err
}
