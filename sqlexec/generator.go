package sqlexec

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-openapi/inflect"
)

// Generator generates a new command.
type Generator struct {
	// FileSystem represents the project directory file system.
	FileSystem FileSystem
}

// Create crates a new file and command for given file name and command name.
func (g *Generator) Create(path, name string) (string, string, error) {
	path = inflect.Underscore(strings.ToLower(path))
	name = inflect.Dasherize(strings.ToLower(name))

	provider := &Provider{}

	if err := provider.ReadDir(g.FileSystem); err != nil {
		return "", "", err
	}

	if _, err := provider.Query(name); err == nil {
		return "", "", fmt.Errorf("Query '%s' already exists", name)
	}

	if path == "" {
		path = time.Now().Format(format)
	}

	path = fmt.Sprintf("%s.sql", path)
	file, err := g.FileSystem.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return "", "", err
	}

	defer file.Close()

	fmt.Fprintln(file, "-- Auto-generated at", time.Now().Format(time.UnixDate))
	fmt.Fprintf(file, "-- name: %s", name)
	fmt.Fprintln(file)
	fmt.Fprintln(file)

	return name, path, err
}
