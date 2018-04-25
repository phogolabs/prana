package sqlmodel

import (
	"bytes"
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
	// ModelGenerator is the SQL model generator
	ModelGenerator Generator
	// QueryGenerator is the SQL script generator
	QueryGenerator Generator
	// Provider provides information the database schema
	Provider SchemaProvider
}

// Write writes the generated schema sqlmodels to a writer
func (e *Executor) Write(w io.Writer, spec *Spec) error {
	_, err := e.writeSchema(w, spec)
	return err
}

// Create creates a package with the generated schema sqlmodels
func (e *Executor) Create(spec *Spec) (string, error) {
	reader := &bytes.Buffer{}
	schema, err := e.writeSchema(reader, spec)
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

	filepath, err := e.modelFileOf(schema, spec)
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

// CreateScript creates a package with the generated schema sqlmodels
func (e *Executor) CreateScript(spec *Spec) (string, error) {
	schema, err := e.schemaOf(spec)
	if err != nil {
		return "", err
	}

	reader := &bytes.Buffer{}
	ctx := &GeneratorContext{
		Writer:  reader,
		Package: e.packageOf(schema, spec),
		Schema:  schema,
	}

	if err = e.QueryGenerator.Generate(ctx); err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	if len(body) == 0 {
		return "", nil
	}

	filepath, err := e.scriptFileOf(schema, spec)
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

func (e *Executor) writeSchema(w io.Writer, spec *Spec) (*Schema, error) {
	schema, err := e.schemaOf(spec)
	if err != nil {
		return nil, err
	}

	ctx := &GeneratorContext{
		Writer:  w,
		Package: e.packageOf(schema, spec),
		Schema:  schema,
	}

	if err = e.ModelGenerator.Generate(ctx); err != nil {
		return nil, err
	}

	return schema, nil
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

func (e *Executor) scriptFileOf(schema *Schema, spec *Spec) (string, error) {
	dir, err := filepath.Abs(spec.Dir)
	if err != nil {
		return "", err
	}

	filename := "command.sql"

	if !schema.IsDefault || e.Config.KeepSchema {
		filename = fmt.Sprintf("%s.sql", schema.Name)
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}

	return filepath.Join(dir, filename), nil
}

func (e *Executor) modelFileOf(schema *Schema, spec *Spec) (string, error) {
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
