package gom

import (
	"io"
	"os"
	"path/filepath"
)

// Dir implements FileSystem using the native file system restricted to a
// specific directory tree.
type Dir string

// MkdirAll creates a directory named path
func (d Dir) MkdirAll(dir string, perm os.FileMode) error {
	dir = filepath.Join(string(d), dir)
	return os.MkdirAll(dir, 0700)
}

// OpenFile is the generalized open call; most users will use Open
func (d Dir) OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	dir := filepath.Join(string(d), filepath.Dir(name))
	name = filepath.Join(dir, filepath.Base(name))
	return os.OpenFile(name, flag, perm)
}

// Walk walks the file tree rooted at root, calling walkFn for each file or
// directory in the tree, including root.
func (d Dir) Walk(dir string, fn filepath.WalkFunc) error {
	dir = filepath.Join(string(d), dir)

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		path, err = filepath.Rel(string(d), path)

		if err != nil {
			return err
		}

		return fn(path, info, err)
	})
}
