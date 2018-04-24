package sqlmodel

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ExecutorConfig defines the Executor's configuration
type ExecutorConfig struct {
	// KeepSchema controlls whether the database schema to be kept as package
	KeepSchema bool
}

// Executor executes the schema generation
type Executor struct {
	// Config of the executor
	Config *ExecutorConfig
	// Composer is the sqlmodel generator
	Composer Composer
	// Provider provides information the database schema
	Provider Provider
}

// Write writes the generated schema sqlmodels to a writer
func (e *Executor) Write(w io.Writer, spec *Spec) error {
	schema, err := e.schemaOf(spec)
	if err != nil {
		return err
	}

	pkg := e.packageOf(schema, spec)
	r, err := e.Composer.Compose(pkg, schema)

	if err != nil {
		return err
	}

	_, err = io.Copy(w, r)
	return err
}

// Create creates a package with the generated schema sqlmodels
func (e *Executor) Create(spec *Spec) (string, error) {
	schema, err := e.schemaOf(spec)
	if err != nil {
		return "", err
	}

	reader, err := e.Composer.Compose(e.packageOf(schema, spec), schema)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	if len(body) == 0 {
		return "", nil
	}

	filepath, err := e.fileOf(schema, spec)
	if err != nil {
		return "", err
	}

	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}

	defer func() {
		if ioErr := file.Close(); err == nil {
			err = ioErr
		}
	}()

	if _, err = file.Write(body); err != nil {
		return "", err
	}

	return filepath, nil
}

func (e *Executor) schemaOf(spec *Spec) (*Schema, error) {
	if len(spec.Tables) == 0 {
		tables, err := e.Provider.Tables(spec.Schema)
		if err != nil {
			return nil, err
		}

		spec.Tables = tables
	}

	schema, err := e.Provider.Schema(spec.Schema, spec.Tables...)
	if err != nil {
		return nil, err
	}

	return schema, nil
}

func (e *Executor) fileOf(schema *Schema, spec *Spec) (string, error) {
	dir, err := filepath.Abs(spec.Dir)
	if err != nil {
		return "", err
	}

	filename := "schema.go"

	if !schema.IsDefault {
		if e.Config.KeepSchema {
			dir = filepath.Join(dir, schema.Name)
		} else {
			filename = fmt.Sprintf("%s.go", schema.Name)
		}
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}

	return filepath.Join(dir, filename), nil
}

func (e *Executor) packageOf(schema *Schema, spec *Spec) string {
	if schema.IsDefault || !e.Config.KeepSchema {
		return filepath.Base(spec.Dir)
	}

	return schema.Name
}
