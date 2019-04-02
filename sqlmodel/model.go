// Package sqlmodel provides primitives for generating structs from database schema
package sqlmodel

import (
	"database/sql"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/phogolabs/parcello"
)

//go:generate parcello -r -i fixture/

var (
	intDef = &TypeDef{
		Type:         "int",
		NullableType: "schema.NullInt",
	}
	uintDef = &TypeDef{
		Type:         "Uint",
		NullableType: "schema.NullUint",
	}
	int16Def = &TypeDef{
		Type:         "int16",
		NullableType: "schema.NullInt16",
	}
	int64Def = &TypeDef{
		Type:         "int64",
		NullableType: "schema.NullInt64",
	}
	int8Def = &TypeDef{
		Type:         "int8",
		NullableType: "schema.NullInt8",
	}
	uint8Def = &TypeDef{
		Type:         "uint8",
		NullableType: "schema.NullUint8",
	}
	uint16Def = &TypeDef{
		Type:         "uint16",
		NullableType: "schema.NullUint16",
	}
	uint32Def = &TypeDef{
		Type:         "uint32",
		NullableType: "schema.NullUint32",
	}
	int32Def = &TypeDef{
		Type:         "int32",
		NullableType: "schema.NullInt32",
	}
	uint64Def = &TypeDef{
		Type:         "uint64",
		NullableType: "schema.NullUint64",
	}
	float32Def = &TypeDef{
		Type:         "float32",
		NullableType: "schema.NullFloat32",
	}
	float64Def = &TypeDef{
		Type:         "float64",
		NullableType: "schema.NullFloat64",
	}
	stringDef = &TypeDef{
		Type:         "string",
		NullableType: "schema.NullString",
	}
	byteDef = &TypeDef{
		Type:         "byte",
		NullableType: "schema.NullByte",
	}
	byteSliceDef = &TypeDef{
		Type:         "[]byte",
		NullableType: "schema.NullBytes",
	}
	boolDef = &TypeDef{
		Type:         "bool",
		NullableType: "schema.NullBool",
	}
	timeDef = &TypeDef{
		Type:         "time.Time",
		NullableType: "schema.NullTime",
	}
	uuidDef = &TypeDef{
		Type:         "schema.UUID",
		NullableType: "schema.NullUUID",
	}
	jsonDef = &TypeDef{
		Type:         "[]byte",
		NullableType: "schema.NullJSON",
	}
	hstoreDef = &TypeDef{
		Type:         "hstore.Hstore",
		NullableType: "hstore.Hstore",
	}
)

//go:generate counterfeiter -fake-name SchemaProvider -o ../fake/SchemaProvider.go . SchemaProvider
//go:generate counterfeiter -fake-name ModelGenerator -o ../fake/ModelGenerator.go . Generator
//go:generate counterfeiter -fake-name TagBuilder -o ../fake/TagBuilder.go . TagBuilder
//go:generate counterfeiter -fake-name Querier -o ../fake/Querier.go . Querier

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
	// PrimaryKeyArgs is the primary key args
	PrimaryKeyArgs string
	// PrimaryKey is the map of primary key args
	PrimaryKey []string
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
type FileSystem = parcello.FileSystem

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
		return t.NullableType
	}
	return t.Type
}

// Spec specifies the generation options
type Spec struct {
	// Filename of the spec
	Filename string
	// FileSystem is the underlying file system
	FileSystem FileSystem
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
