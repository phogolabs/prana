package sqlmodel

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/go-openapi/inflect"
	"golang.org/x/tools/imports"
)

var _ ModelGenerator = &Generator{}

// GeneratorConfig controls how the code generation happens
type GeneratorConfig struct {
	// KeepSchema controlls whether the database schema to be kept as package
	KeepSchema bool
	// InlcudeDoc determines whether to include documentation
	InlcudeDoc bool
	// IgnoreTables ecludes the those tables from generation
	IgnoreTables []string
}

// Generator generates Golang structs from database schema
type Generator struct {
	// TagBuilder builds struct tags from column type
	TagBuilder TagBuilder
	// Config controls how the code generation happens
	Config *GeneratorConfig
}

// GenerateModel generates the golang structs from database schema
func (g *Generator) GenerateModel(pkg string, schema *Schema) (io.Reader, error) {
	buffer := &bytes.Buffer{}
	tables := g.tables(schema)

	if len(tables) == 0 {
		return buffer, nil
	}

	g.writePackage(pkg, schema.Name, buffer)

	for _, table := range tables {
		g.writeTable(pkg, schema.IsDefault, &table, buffer)
	}

	if err := g.format(buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}

// GenerateSQLScript generates a script for given schema
func (g *Generator) GenerateSQLScript(schema *Schema) (io.Reader, error) {
	buffer := &bytes.Buffer{}
	g.writeSQLComment(buffer)

	tables := g.tables(schema)

	for _, table := range tables {
		g.writeSQLQuerySelectAll(buffer, schema, &table)
		g.writeSQLQuerySelect(buffer, schema, &table)
		g.writeSQLQueryInsert(buffer, schema, &table)
		g.writeSQLQueryUpdate(buffer, schema, &table)
		g.writeSQLQueryDelete(buffer, schema, &table)
	}

	return buffer, nil
}

func (g *Generator) tables(schema *Schema) []Table {
	tables := []Table{}
	ignore := g.ignore()

	for _, table := range schema.Tables {
		if index := sort.SearchStrings(ignore, table.Name); index >= 0 && index < len(ignore) {
			continue
		}

		tables = append(tables, table)
	}
	return tables
}

func (g *Generator) writePackage(pkg, name string, buffer io.Writer) {
	if g.Config.InlcudeDoc {
		fmt.Fprintf(buffer, "// Package %s contains an object model of database schema '%s'", pkg, name)
		fmt.Fprintln(buffer)
		fmt.Fprintln(buffer, "// Auto-generated at", time.Now().Format(time.UnixDate))
	}

	fmt.Fprintf(buffer, "package ")
	fmt.Fprintf(buffer, pkg)
	fmt.Fprintln(buffer)
}

func (g *Generator) writeTable(pkg string, isDefaultSchema bool, table *Table, buffer io.Writer) {
	columns := table.Columns
	length := len(columns)
	typeName := g.typeName(pkg, isDefaultSchema, table)

	if g.Config.InlcudeDoc {
		fmt.Fprintln(buffer)
		fmt.Fprintf(buffer, "// %s represents a data base table '%s'", typeName, table.Name)
		fmt.Fprintln(buffer)
	}

	fmt.Fprintf(buffer, "type %v struct {", typeName)
	fmt.Fprintln(buffer)

	for index, column := range columns {
		current := column
		fieldName := g.fieldName(&current)
		fieldType := g.fieldType(&current)
		fieldTag := g.TagBuilder.Build(&current)

		if g.Config.InlcudeDoc {
			if index > 0 {
				fmt.Fprintln(buffer)
			}
			fmt.Fprintf(buffer, "// %s represents a database column '%s' of type '%v'", fieldName, column.Name, column.Type)
			fmt.Fprintln(buffer)
		}

		fmt.Fprint(buffer, fieldName)
		fmt.Fprint(buffer, " ")
		fmt.Fprint(buffer, fieldType)
		fmt.Fprint(buffer, " ")
		fmt.Fprint(buffer, fieldTag)

		fmt.Fprintln(buffer)

		if index == length-1 {
			fmt.Fprintln(buffer, "}")
		}
	}
}

func (g *Generator) ignore() []string {
	ignore := g.Config.IgnoreTables

	if !sort.StringsAreSorted(ignore) {
		sort.Strings(ignore)
	}

	return ignore
}

func (g *Generator) typeName(pkg string, isDefaultSchema bool, table *Table) string {
	name := inflect.Camelize(table.Name)
	name = inflect.Singularize(name)

	if !g.Config.KeepSchema && !isDefaultSchema {
		pkg = inflect.Camelize(pkg)
		name = fmt.Sprintf("%s%s", pkg, name)
	}

	return name
}

func (g *Generator) fieldName(column *Column) string {
	name := inflect.Camelize(column.Name)
	name = strings.Replace(name, "Id", "ID", -1)
	return name
}

func (g *Generator) fieldType(column *Column) string {
	return column.ScanType
}

func (g *Generator) format(buffer *bytes.Buffer) error {
	data, err := imports.Process("model", buffer.Bytes(), nil)
	if err != nil {
		return err
	}

	data, err = format.Source(data)
	if err != nil {
		return err
	}

	buffer.Reset()

	_, err = buffer.Write(data)
	return err
}

func (g *Generator) writeSQLQuerySelectAll(w io.Writer, schema *Schema, table *Table) {
	tableName := g.tableName(schema, table)
	fmt.Fprintf(w, "-- name: select-all-%s\n", g.commandName(tableName, false))
	fmt.Fprintf(w, "SELECT * FROM %s\n\n", tableName)
}

func (g *Generator) writeSQLQuerySelect(w io.Writer, schema *Schema, table *Table) {
	tableName := g.tableName(schema, table)
	fmt.Fprintf(w, "-- name: select-%s\n", g.commandName(tableName, true))
	fmt.Fprintf(w, "SELECT * FROM %s\n", tableName)
	fmt.Fprintf(w, "WHERE %s\n\n", g.pkCondition(table))
}

func (g *Generator) writeSQLQueryInsert(w io.Writer, schema *Schema, table *Table) {
	tableName := g.tableName(schema, table)
	columns, values := g.insertParam(table)
	fmt.Fprintf(w, "-- name: insert-%s\n", g.commandName(tableName, true))
	fmt.Fprintf(w, "INSERT INTO %s (%s)\n", tableName, columns)
	fmt.Fprintf(w, "VALUES (%s)\n\n", values)
}

func (g *Generator) writeSQLQueryUpdate(w io.Writer, schema *Schema, table *Table) {
	tableName := g.tableName(schema, table)
	condition, values := g.updateParam(table)
	fmt.Fprintf(w, "-- name: update-%s\n", g.commandName(tableName, true))
	fmt.Fprintf(w, "UPDATE %s\n", tableName)
	fmt.Fprintf(w, "SET %s\n", values)
	fmt.Fprintf(w, "WHERE %s\n\n", condition)
}

func (g *Generator) writeSQLQueryDelete(w io.Writer, schema *Schema, table *Table) {
	tableName := g.tableName(schema, table)
	fmt.Fprintf(w, "-- name: delete-%s\n", g.commandName(tableName, true))
	fmt.Fprintf(w, "DELETE FROM %s\n", tableName)
	fmt.Fprintf(w, "WHERE %s\n\n", g.pkCondition(table))
}

func (g *Generator) writeSQLComment(w io.Writer) {
	if g.Config.InlcudeDoc {
		fmt.Fprintln(w, "-- Auto-generated at", time.Now().Format(time.UnixDate))
		fmt.Fprintln(w)
	}
}

func (g *Generator) commandName(name string, singularize bool) string {
	name = strings.Replace(name, ".", "-", -1)
	if singularize {
		name = inflect.Singularize(name)
	}
	return name
}

func (g *Generator) tableName(schema *Schema, table *Table) string {
	name := table.Name

	if !schema.IsDefault || g.Config.KeepSchema {
		name = fmt.Sprintf("%s.%s", schema.Name, name)
	}

	return name
}

func (g *Generator) insertParam(table *Table) (string, string) {
	columns := []string{}
	values := []string{}

	for _, column := range table.Columns {
		columns = append(columns, column.Name)
		values = append(values, "?")
	}

	return strings.Join(columns, ", "), strings.Join(values, ", ")
}

func (g *Generator) updateParam(table *Table) (string, string) {
	values := []string{}
	conditions := []string{}

	for _, column := range table.Columns {
		if column.Type.IsPrimaryKey {
			conditions = append(conditions, fmt.Sprintf("%s = ?", column.Name))
			continue
		}
		values = append(values, fmt.Sprintf("%s = ?", column.Name))
	}

	return strings.Join(conditions, ", "), strings.Join(values, ", ")
}

func (g *Generator) pkCondition(table *Table) string {
	conditions := []string{}

	for _, column := range table.Columns {
		if !column.Type.IsPrimaryKey {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s = ?", column.Name))
	}

	return strings.Join(conditions, " AND ")
}
