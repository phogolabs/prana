// Package sqlmodel provides primitives for generating structs from database schema
package sqlmodel

import (
	"database/sql"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"
)

//go:embed template/*
var template embed.FS

var (
	intDef = &TypeDef{
		Type: "int",
	}
	uintDef = &TypeDef{
		Type: "Uint",
	}
	int16Def = &TypeDef{
		Type: "int16",
	}
	int64Def = &TypeDef{
		Type: "int64",
	}
	int8Def = &TypeDef{
		Type: "int8",
	}
	uint8Def = &TypeDef{
		Type: "uint8",
	}
	uint16Def = &TypeDef{
		Type: "uint16",
	}
	uint32Def = &TypeDef{
		Type: "uint32",
	}
	int32Def = &TypeDef{
		Type: "int32",
	}
	uint64Def = &TypeDef{
		Type: "uint64",
	}
	float32Def = &TypeDef{
		Type: "float32",
	}
	float64Def = &TypeDef{
		Type: "float64",
	}
	stringDef = &TypeDef{
		Type: "string",
	}
	byteDef = &TypeDef{
		Type: "byte",
	}
	byteSliceDef = &TypeDef{
		Type:         "[]byte",
		NullableType: "[]byte",
	}
	boolDef = &TypeDef{
		Type: "bool",
	}
	timeDef = &TypeDef{
		Type: "time.Time",
	}
	uuidDef = &TypeDef{
		Type: "schema.UUID",
	}
	jsonDef = &TypeDef{
		Type:         "[]byte",
		NullableType: "[]byte",
	}
	hstoreDef = &TypeDef{
		Type:         "hstore.Hstore",
		NullableType: "hstore.Hstore",
	}
)

//go:generate counterfeiter -fake-name SchemaProvider -o ../fake/schema_provider.go . SchemaProvider
//go:generate counterfeiter -fake-name ModelGenerator -o ../fake/model_generator.go . Generator
//go:generate counterfeiter -fake-name TagBuilder -o ../fake/tag_builder.go . TagBuilder
//go:generate counterfeiter -fake-name Querier -o ../fake/querier.go . Querier

// Querier executes queries
type Querier interface {
	// Query performs a query and returns a set of rows
	Query(query string, args ...interface{}) (*sql.Rows, error)
	// QueryRow performs a query and returns a row
	QueryRow(query string, args ...interface{}) *sql.Row
	// Close closes the connection
	Close() error
}

// SchemaProvider provides a metadata for database schema
type SchemaProvider interface {
	// Tables returns all tables for this schema
	Tables(schema string) ([]string, error)
	// Schema returns the schema definition
	Schema(schema string, tables ...string) (*Schema, error)
	// Close closes connection to the db
	Close() error
}

// GeneratorContext is the generator's context
type GeneratorContext struct {
	// Template name
	Template string
	// Writer where the output will be written
	Writer io.Writer
	// Schema definition
	Schema *Schema
}

// Generator generates the sqlmodels
type Generator interface {
	// Generate generates a model or script
	Generate(ctx *GeneratorContext) error
}

// Schema represents a database schema
type Schema struct {
	// Name of the schema
	Name string
	// Driver name
	Driver string
	// Tables are the associated tables
	Tables []Table
	// IsDefault returns if this schema is default
	IsDefault bool
	// Model for this schema
	Model SchemaModel
}

// SchemaModel represents the schema's model
type SchemaModel struct {
	// Package name
	Package string
	// HasDocumentation return true if the schema has documentation
	HasDocumentation bool
}

// Table represents a table name and its schema
type Table struct {
	// Name of this table
	Name string
	// Driver name
	Driver string
	// Model representation of this table
	Model TableModel
	// Columns of this table
	Columns []Column
}

// TableModel represents the model definition
type TableModel struct {
	// HasDocumentation return true if the table has documentation
	HasDocumentation bool
	// Type of this model
	Type string
	// Package name
	Package string
	// InsertRoutine is the insert routine name
	InsertRoutine string
	// InsertColumns are the columns
	InsertColumns string
	// InsertValues are the values to be inserted
	InsertValues string
	// SelectByPKRoutine is the select by primary key routine
	SelectByPKRoutine string
	// SelectAllRoutine is the select-all's routine
	SelectAllRoutine string
	// DeleteByPkRoutine is the delete by primary key routine
	DeleteByPKRoutine string
	// UpdateByPKRoutine is the update by primary key routine
	UpdateByPKRoutine string
	// UpdateByPKColumns is the columns for update condition
	UpdateByPKColumns string
	// PrimaryKeyCondition is a where clause condition
	PrimaryKeyCondition string
	// PrimaryKeyParams is the primary key args
	PrimaryKeyParams string
	// PrimaryKeyEntityParams is the primary key args
	PrimaryKeyEntityParams string
	// PrimaryKeyArgs is the primary key args
	PrimaryKeyArgs string
	// PrimaryKey is the map of primary key args
	PrimaryKey map[string]string
}

// Column represents a metadata for database column
type Column struct {
	// Name is the name of this column
	Name string
	// Type is the database type of this column
	Type ColumnType
	// ScanType is the database type of this column
	ScanType string
	// Model representation of this column
	Model ColumnModel
}

// ColumnModel represents the field definition for given column
type ColumnModel struct {
	// HasDocumentation return true if the column has documentation
	HasDocumentation bool
	// Name is the name of this column
	Name string
	// Type is the database type of this column
	Type string
	// Tage is the field tag
	Tag string
}

// ColumnType is the type of the column
type ColumnType struct {
	// Name of the column type
	Name string
	// Underlying is the name of the column data type (the underlying type of the domain, if applicable)
	Underlying string
	// IsPrimaryKey returns true if the column is in primary key
	IsPrimaryKey bool
	// IsNullable determines whether the column allow null values
	IsNullable bool
	// IsUnsigned returns true if the numeric type is unassigned
	IsUnsigned bool
	// CharMaxLength determines the maximum length for character types
	CharMaxLength int
	// Precision for numeric type
	Precision int
	// PrecisionScale for numeric type
	PrecisionScale int
}

// DBType returns the db type as string
func (t ColumnType) DBType() string {
	name := t.Name

	if strings.EqualFold(name, "USER-DEFINED") {
		name = t.Underlying
	}

	if t.CharMaxLength > 0 {
		name = fmt.Sprintf("%s(%d)", name, t.CharMaxLength)
	} else if t.Precision > 0 && t.PrecisionScale == 0 {
		name = fmt.Sprintf("%s(%d)", name, t.Precision)
	} else if t.Precision > 0 && t.PrecisionScale > 0 {
		name = fmt.Sprintf("%s(%d, %d)", name, t.Precision, t.PrecisionScale)
	}

	return name
}

// String represents the ColumnType as string
func (t ColumnType) String() string {
	name := t.DBType()

	if t.IsPrimaryKey {
		name = fmt.Sprintf("%s PRIMARY KEY", name)
	}

	if t.IsNullable {
		name = fmt.Sprintf("%s NULL", name)
	} else {
		name = fmt.Sprintf("%s NOT NULL", name)
	}

	return strings.ToUpper(name)
}

// FileSystem provides with primitives to work with the underlying file system
type FileSystem = fs.FS

// WriteFileSystem represents a wriable file system
type WriteFileSystem interface {
	FileSystem

	// OpenFile opens a new file
	OpenFile(string, int, fs.FileMode) (fs.File, error)
}

// TypeDef represents a type definition
type TypeDef struct {
	// Type name
	Type string
	// NullableType name
	NullableType string
}

// As returns the type name if nullable is true, otherwise the nullable type
func (t *TypeDef) As(nullable bool) string {
	if nullable {
		if t.NullableType != "" {
			return t.NullableType
		}

		return "*" + t.Type
	}

	return t.Type
}

// Spec specifies the generation options
type Spec struct {
	// Filename of the spec
	Filename string
	// Template name
	Template string
	// FileSystem is the underlying file system
	FileSystem WriteFileSystem
	// Schema is the database schema name
	Schema string
	// Tables is the list of the desired tables from the database schema
	Tables []string
	// IgnoreTables ecludes the those tables from generation
	IgnoreTables []string
}

// TagBuilder builds tags from column type
type TagBuilder interface {
	// Build returns a struct tag from column type
	Build(column *Column) string
}

type sqliteInf struct {
	CID          int
	Type         string
	NotNullable  int
	DefaultValue interface{}
	PK           int
}

func contains(arr []string, item string) bool {
	if index := sort.SearchStrings(arr, item); index < len(arr) {
		return arr[index] == item
	}

	return false
}
