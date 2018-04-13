// Package shema provides primitives for generating structs from database schema
package schema

import (
	"fmt"
	"io"
	"strings"
)

var (
	IntDef = &TypeDef{
		Type:         "int",
		NullableType: "null.Int",
	}
	UIntDef = &TypeDef{
		Type:         "Uint",
		NullableType: "null.Uint",
	}
	Int16Def = &TypeDef{
		Type:         "int16",
		NullableType: "null.Int16",
	}
	Int64Def = &TypeDef{
		Type:         "int64",
		NullableType: "null.Int64",
	}
	Int8Def = &TypeDef{
		Type:         "int8",
		NullableType: "null.Int8",
	}
	UInt8Def = &TypeDef{
		Type:         "uint8",
		NullableType: "null.Uint8",
	}
	UInt16Def = &TypeDef{
		Type:         "uint16",
		NullableType: "null.Uint16",
	}
	UInt32Def = &TypeDef{
		Type:         "uint32",
		NullableType: "null.Uint32",
	}
	Int32Def = &TypeDef{
		Type:         "int32",
		NullableType: "null.Int32",
	}
	UInt64Def = &TypeDef{
		Type:         "uint64",
		NullableType: "null.Uint64",
	}
	Float32Def = &TypeDef{
		Type:         "float32",
		NullableType: "null.Float32",
	}
	Float64Def = &TypeDef{
		Type:         "float64",
		NullableType: "null.Float64",
	}
	StringDef = &TypeDef{
		Type:         "string",
		NullableType: "null.String",
	}
	ByteDef = &TypeDef{
		Type:         "byte",
		NullableType: "null.Byte",
	}
	ByteSliceDef = &TypeDef{
		Type:         "[]byte",
		NullableType: "null.Bytes",
	}
	BoolDef = &TypeDef{
		Type:         "bool",
		NullableType: "null.Bool",
	}
	TimeDef = &TypeDef{
		Type:         "time.Time",
		NullableType: "null.Time",
	}
	UUIDDef = &TypeDef{
		Type:         "uuid.UUID",
		NullableType: "uuid.NullUUID",
	}
	JSONDef = &TypeDef{
		Type:         "[]byte",
		NullableType: "null.JSON",
	}
	HStoreDef = &TypeDef{
		Type:         "hstore.Hstore",
		NullableType: "hstore.Hstore",
	}
)

//go:generate counterfeiter -fake-name SchemaProvider -o ../fake/SchemaProvider.go . Provider
//go:generate counterfeiter -fake-name SchemaComposer -o ../fake/SchemaComposer.go . Composer

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

// String represents the ColumnType as string
func (t ColumnType) String() string {
	name := t.Name

	if t.CharMaxLength > 0 {
		name = fmt.Sprintf("%s(%d)", name, t.CharMaxLength)
	} else if t.Precision > 0 && t.PrecisionScale == 0 {
		name = fmt.Sprintf("%s(%d)", name, t.Precision)
	} else if t.Precision > 0 && t.PrecisionScale > 0 {
		name = fmt.Sprintf("%s(%d, %d)", name, t.Precision, t.PrecisionScale)
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
