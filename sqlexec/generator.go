package sqlexec

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/go-openapi/inflect"
)

// Generator generates a new command.
type Generator struct {
	// FileSystem represents the project directory file system.
	FileSystem WriteFileSystem
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

	now := time.Now().UTC()

	if path == "" {
		path = now.Format(format)
	}

	path = fmt.Sprintf("%s.sql", path)

	file, err := g.FileSystem.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return "", "", err
	}

	defer file.Close()

	if writer, ok := file.(io.Writer); ok {
		fmt.Fprintf(writer, "-- name: %s", name)
		fmt.Fprintln(writer)
		fmt.Fprintln(writer)
	}

	return name, path, err
}
