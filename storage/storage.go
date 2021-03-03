package storage

import (
	"io/fs"
	"os"
	"path/filepath"
)

type (
	File     = fs.File
	FileMode = fs.FileMode
)

// FileSystem represents a disk file system
type FileSystem struct {
	dir string
}

// NewStorage creates a new storage
func New(dir string) *FileSystem {
	return &FileSystem{dir: dir}
}

// Open opens the named file for reading
func (storage *FileSystem) Open(name string) (File, error) {
	name = filepath.Join(storage.dir, name)

	if err := storage.mkdir(name); err != nil {
		return nil, err
	}

	return os.Open(name)
}

// OpenFile is the generalized open call
func (storage *FileSystem) OpenFile(name string, flag int, perm FileMode) (File, error) {
	name = filepath.Join(storage.dir, name)

	if err := storage.mkdir(name); err != nil {
		return nil, err
	}

	return os.OpenFile(name, flag, perm)
}

func (storage *FileSystem) mkdir(name string) error {
	if path := filepath.Dir(name); path != "" {
		return os.MkdirAll(path, 0700)
	}

	return nil
}
