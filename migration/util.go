package migration

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
)

//go:generate counterfeiter -fake-name MigrationFileSystem -o ../fake/MigrationFileSystem.go . FileSystem

// FileSystem provides with primitives to work with the underlying file system
type FileSystem interface {
	// Walk walks the file tree rooted at root, calling walkFn for each file or
	// directory in the tree, including root.
	Walk(fn filepath.WalkFunc) error
	// Open opens the named file for reading.
	Open(path string) (io.ReadCloser, error)
	// WriteFile writes data to a file named by filename.
	WriteFile(filename string, data []byte, perm os.FileMode) error
	// Join joins any number of path elements into a single path
	Join(elem ...string) string
}

// Dir implements FileSystem using the native file system restricted to a
// specific directory tree.
type Dir string

// Open opens the named file for reading.
func (d Dir) Open(path string) (io.ReadCloser, error) {
	path = filepath.Join(string(d), path)
	return os.Open(path)
}

// Walk walks the file tree rooted at root, calling walkFn for each file or
// directory in the tree, including root.
func (d Dir) Walk(fn filepath.WalkFunc) error {
	return filepath.Walk(string(d), fn)
}

// WriteFile writes data to a file named by filename.
func (d Dir) WriteFile(filename string, data []byte, perm os.FileMode) error {
	dir := string(d)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	filename = filepath.Join(dir, filename)
	return ioutil.WriteFile(filename, data, perm)
}

// Join joins any number of path elements into a single path
func (d Dir) Join(elem ...string) string {
	path := []string{string(d)}
	path = append(path, elem...)
	return filepath.Join(path...)
}

// RunAll runs all migrations
func RunAll(db *sqlx.DB, fileSystem FileSystem) error {
	executor := &Executor{
		Provider: &Provider{
			FileSystem: fileSystem,
			DB:         db,
		},
		Runner: &Runner{
			FileSystem: fileSystem,
			DB:         db,
		},
		Generator: &Generator{
			FileSystem: fileSystem,
		},
	}

	if err := executor.Setup(); err != nil {
		return err
	}

	_, err := executor.RunAll()
	return err
}
