// Package schema provides primitives for generating structs from database schema
package schema

import (
	"fmt"
	"io"
	"strings"
)

var (
	intDef = &TypeDef{
		Type:         "int",
		NullableType: "null.Int",
	}
	uintDef = &TypeDef{
		Type:         "Uint",
		NullableType: "null.Uint",
	}
	int16Def = &TypeDef{
		Type:         "int16",
		NullableType: "null.Int16",
	}
	int64Def = &TypeDef{
		Type:         "int64",
		NullableType: "null.Int64",
	}
	int8Def = &TypeDef{
		Type:         "int8",
		NullableType: "null.Int8",
	}
	uint8Def = &TypeDef{
		Type:         "uint8",
		NullableType: "null.Uint8",
	}
	uint16Def = &TypeDef{
		Type:         "uint16",
		NullableType: "null.Uint16",
	}
	uint32Def = &TypeDef{
		Type:         "uint32",
		NullableType: "null.Uint32",
	}
	int32Def = &TypeDef{
		Type:         "int32",
		NullableType: "null.Int32",
	}
	uint64Def = &TypeDef{
		Type:         "uint64",
		NullableType: "null.Uint64",
	}
	float32Def = &TypeDef{
		Type:         "float32",
		NullableType: "null.Float32",
	}
	float64Def = &TypeDef{
		Type:         "float64",
		NullableType: "null.Float64",
	}
	stringDef = &TypeDef{
		Type:         "string",
		NullableType: "null.String",
	}
	byteDef = &TypeDef{
		Type:         "byte",
		NullableType: "null.Byte",
	}
	byteSliceDef = &TypeDef{
		Type:         "[]byte",
		NullableType: "null.Bytes",
	}
	boolDef = &TypeDef{
		Type:         "bool",
		NullableType: "null.Bool",
	}
	timeDef = &TypeDef{
		Type:         "time.Time",
		NullableType: "null.Time",
	}
	uuidDef = &TypeDef{
		Type:         "uuid.UUID",
		NullableType: "uuid.NullUUID",
	}
	jsonDef = &TypeDef{
		Type:         "[]byte",
		NullableType: "null.JSON",
	}
	hstoreDef = &TypeDef{
		Type:         "hstore.Hstore",
		NullableType: "hstore.Hstore",
	}
)

//go:generate counterfeiter -fake-name SchemaProvider -o ../fake/SchemaProvider.go . Provider
//go:generate counterfeiter -fake-name SchemaComposer -o ../fake/SchemaComposer.go . Composer
//go:generate counterfeiter -fake-name SchemaTagBuilder -o ../fake/SchemaTagBuilder.go . TagBuilder

// Provider provides a metadata for database schema
type Provider interface {
	// Tables returns all tables for this schema
	Tables(schema string) ([]string, error)
	// Schema returns the schema definition
	Schema(schema string, tables ...string) (*Schema, error)
}

// Composer composes the models
type Composer interface {
	// Compose generates the golang structs from database schema
	Compose(pkg string, sch *Schema) (io.Reader, error)
}

// Schema represents a database schema
type Schema struct {
	// Name of the schema
	Name string
	// Tables are the associated tables
	Tables []Table
	// IsDefault returns if this schema is default
	IsDefault bool
}

// Table represents a table name and its schema
type Table struct {
	// Name of this table
	Name string
	// Columns of this table
	Columns []Column
}

// Column represents a metadata for database column
type Column struct {
	// Name is the name of this column
	Name string
	// Type is the database type of this column
	Type ColumnType
	// ScanType is the scannable data type for this column
	ScanType string
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
	// Schema is the database schema name
	Schema string
	// Tables is the list of the desired tables from the database schema
	Tables []string
	// Dir is a path to root model package directory
	Dir string
}

// TagBuilder builds tags from column type
type TagBuilder interface {
	// Build returns a struct tag from column type
	Build(column *Column) string
}

// FieldTag represents a field tag
type FieldTag struct {
	// Name of the tag
	Name string
	// Options
	Options []string
}

// AddOption adds option to that tag
func (t *FieldTag) AddOption(opt string) {
	t.Options = append(t.Options, opt)
}

// String returns the tag as string
func (t *FieldTag) String() string {
	options := t.Options
	if len(options) == 0 {
		options = append(options, "-")
	}

	return fmt.Sprintf("%s:\"%s\"", t.Name, strings.Join(options, ","))
}

// FieldTagList represents tag of list
type FieldTagList []*FieldTag

// String returns the list as string
func (l FieldTagList) String() string {
	tags := []string{}

	for _, t := range l {
		tags = append(tags, t.String())
	}

	return fmt.Sprintf("`%s`", strings.Join(tags, " "))
}

type sqliteInf struct {
	CID          int
	Type         string
	NotNullable  int
	DefaultValue interface{}
	PK           int
}
