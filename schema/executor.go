package schema

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Executor executes the schema generation
type Executor struct {
	// Composer is the model generator
	Composer Composer
	// Provider provides information the database schema
	Provider Provider
}

// WriteTo writes the generated schema models to a writer
func (e *Executor) WriteTo(w io.Writer, spec *Spec) error {
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

// Create creates a package with the generated schema models
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

	defer file.Close()

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

	if !schema.IsDefault {
		dir = filepath.Join(dir, schema.Name)
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}

	return filepath.Join(dir, "schema.go"), nil
}

func (e *Executor) packageOf(schema *Schema, spec *Spec) string {
	if schema.IsDefault {
		return filepath.Base(spec.Dir)
	}

	return schema.Name
}
