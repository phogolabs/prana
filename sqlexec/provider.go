package sqlexec

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Provider loads SQL sqlexecs and provides all SQL statements as commands.
type Provider struct {
	// Driver is the current SQL driver
	Driver string
	// private fields
	mu         sync.RWMutex
	repository map[string]string
}

// ReadDir loads all sqlexec commands from a given directory. Note that all
// sqlexecs should have .sql extension.
func (p *Provider) ReadDir(fs FileSystem) error {
	return fs.Walk("/", func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return os.ErrNotExist
		}

		if info.IsDir() {
			return nil
		}

		return p.ReadFile(path, fs)
	})
}

// ReadFile reads a given file
func (p *Provider) ReadFile(path string, fs FileSystem) error {
	if !p.filter(path) {
		return nil
	}

	file, err := fs.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	defer func() {
		if ioErr := file.Close(); err == nil {
			err = ioErr
		}
	}()

	if _, err = p.ReadFrom(file); err != nil {
		return err
	}

	return nil
}

// ReadFrom reads the sqlexec from a reader
func (p *Provider) ReadFrom(r io.Reader) (int64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.repository == nil {
		p.repository = make(map[string]string)
	}

	scanner := &Scanner{}
	stmts := scanner.Scan(r)

	for name, stmt := range stmts {
		if _, ok := p.repository[name]; ok {
			return 0, fmt.Errorf("query '%s' already exists", name)
		}

		p.repository[name] = stmt
	}

	return int64(len(stmts)), nil
}

// Query returns a query statement for given name and parameters. The operation can
// err if the command cannot be found.
func (p *Provider) Query(name string, params ...Param) (Query, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if query, ok := p.repository[name]; ok {
		return &Stmt{
			query:  query,
			params: params,
		}, nil
	}

	return nil, nonExistQueryErr(name)
}

// NamedQuery returns a query statement for given name and parameters. The operation can
// err if the command cannot be found.
func (p *Provider) NamedQuery(name string, param Param) (Query, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if query, ok := p.repository[name]; ok {
		return &NamedStmt{
			query: query,
			param: param,
		}, nil
	}

	return nil, nonExistQueryErr(name)
}

// Filter returns true if the file can be processed for the current driver
func (p *Provider) filter(path string) bool {
	ext := filepath.Ext(path)

	if ext != ".sql" {
		return false
	}

	driver := PathDriver(path)
	return driver == p.Driver
}

// PathDriver returns the driver name from a given path
func PathDriver(path string) string {
	ext := filepath.Ext(path)
	_, path = filepath.Split(path)
	path = strings.Replace(path, ext, "", -1)
	parts := strings.Split(path, "_")
	driver := strings.ToLower(parts[len(parts)-1])

	switch driver {
	case "sqlite3", "postgres", "mysql":
		return driver
	default:
		return ""
	}
}

func nonExistQueryErr(name string) error {
	return fmt.Errorf("query '%s' not found", name)
}
