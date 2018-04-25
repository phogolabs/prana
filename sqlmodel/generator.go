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

var _ Generator = &ModelGenerator{}
var _ Generator = &QueryGenerator{}

// ModelGenerator generates Golang structs from database schema
type ModelGenerator struct {
	// TagBuilder builds struct tags from column type
	TagBuilder TagBuilder
	// Config controls how the code generation happens
	Config *GeneratorConfig
}

// Generate generates the golang structs from database schema
func (g *ModelGenerator) Generate(ctx *GeneratorContext) error {
	pkg := ctx.Package
	schema := ctx.Schema
	buffer := &bytes.Buffer{}
	tables := tables(g.Config.IgnoreTables, schema)

	if len(tables) == 0 {
		return nil
	}

	g.writePackage(pkg, schema.Name, buffer)

	for _, table := range tables {
		g.writeTable(pkg, schema.IsDefault, &table, buffer)
	}

	if err := g.format(buffer); err != nil {
		return err
	}

	_, err := io.Copy(ctx.Writer, buffer)
	return err
}

func (g *ModelGenerator) writePackage(pkg, name string, buffer io.Writer) {
	if g.Config.InlcudeDoc {
		fmt.Fprintf(buffer, "// Package %s contains an object model of database schema '%s'", pkg, name)
		fmt.Fprintln(buffer)
		fmt.Fprintln(buffer, "// Auto-generated at", time.Now().Format(time.UnixDate))
	}

	fmt.Fprintf(buffer, "package ")
	fmt.Fprintf(buffer, pkg)
	fmt.Fprintln(buffer)
}

func (g *ModelGenerator) writeTable(pkg string, isDefaultSchema bool, table *Table, buffer io.Writer) {
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

func (g *ModelGenerator) typeName(pkg string, isDefaultSchema bool, table *Table) string {
	name := inflect.Camelize(table.Name)
	name = inflect.Singularize(name)

	if !g.Config.KeepSchema && !isDefaultSchema {
		pkg = inflect.Camelize(pkg)
		name = fmt.Sprintf("%s%s", pkg, name)
	}

	return name
}

func (g *ModelGenerator) fieldName(column *Column) string {
	name := inflect.Camelize(column.Name)
	name = strings.Replace(name, "Id", "ID", -1)
	return name
}

func (g *ModelGenerator) fieldType(column *Column) string {
	return column.ScanType
}

func (g *ModelGenerator) format(buffer *bytes.Buffer) error {
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

// QueryGenerator generates queries for give schema
type QueryGenerator struct {
	// Config controls how the code generation happens
	Config *GeneratorConfig
}

// Generate generates a script for given schema
func (g *QueryGenerator) Generate(ctx *GeneratorContext) error {
	schema := ctx.Schema
	buffer := &bytes.Buffer{}
	g.writeSQLComment(buffer)

	tables := tables(g.Config.IgnoreTables, schema)

	for _, table := range tables {
		g.writeSQLQuerySelectAll(buffer, schema, &table)
		g.writeSQLQuerySelect(buffer, schema, &table)
		g.writeSQLQueryInsert(buffer, schema, &table)
		g.writeSQLQueryUpdate(buffer, schema, &table)
		g.writeSQLQueryDelete(buffer, schema, &table)
	}

	_, err := io.Copy(ctx.Writer, buffer)
	return err
}

func (g *QueryGenerator) writeSQLQuerySelectAll(w io.Writer, schema *Schema, table *Table) {
	tableName := g.tableName(schema, table)
	fmt.Fprintf(w, "-- name: select-all-%s\n", g.commandName(tableName, false))
	fmt.Fprintf(w, "SELECT * FROM %s\n\n", tableName)
}

func (g *QueryGenerator) writeSQLQuerySelect(w io.Writer, schema *Schema, table *Table) {
	tableName := g.tableName(schema, table)
	fmt.Fprintf(w, "-- name: select-%s\n", g.commandName(tableName, true))
	fmt.Fprintf(w, "SELECT * FROM %s\n", tableName)
	fmt.Fprintf(w, "WHERE %s\n\n", g.pkCondition(table))
}

func (g *QueryGenerator) writeSQLQueryInsert(w io.Writer, schema *Schema, table *Table) {
	tableName := g.tableName(schema, table)
	columns, values := g.insertParam(table)
	fmt.Fprintf(w, "-- name: insert-%s\n", g.commandName(tableName, true))
	fmt.Fprintf(w, "INSERT INTO %s (%s)\n", tableName, columns)
	fmt.Fprintf(w, "VALUES (%s)\n\n", values)
}

func (g *QueryGenerator) writeSQLQueryUpdate(w io.Writer, schema *Schema, table *Table) {
	tableName := g.tableName(schema, table)
	condition, values := g.updateParam(table)
	fmt.Fprintf(w, "-- name: update-%s\n", g.commandName(tableName, true))
	fmt.Fprintf(w, "UPDATE %s\n", tableName)
	fmt.Fprintf(w, "SET %s\n", values)
	fmt.Fprintf(w, "WHERE %s\n\n", condition)
}

func (g *QueryGenerator) writeSQLQueryDelete(w io.Writer, schema *Schema, table *Table) {
	tableName := g.tableName(schema, table)
	fmt.Fprintf(w, "-- name: delete-%s\n", g.commandName(tableName, true))
	fmt.Fprintf(w, "DELETE FROM %s\n", tableName)
	fmt.Fprintf(w, "WHERE %s", g.pkCondition(table))
}

func (g *QueryGenerator) writeSQLComment(w io.Writer) {
	if g.Config.InlcudeDoc {
		fmt.Fprintln(w, "-- Auto-generated at", time.Now().Format(time.UnixDate))
		fmt.Fprintln(w)
	}
}

func (g *QueryGenerator) commandName(name string, singularize bool) string {
	name = strings.Replace(name, ".", "-", -1)
	if singularize {
		name = inflect.Singularize(name)
	}
	return name
}

func (g *QueryGenerator) tableName(schema *Schema, table *Table) string {
	name := table.Name

	if !schema.IsDefault || g.Config.KeepSchema {
		name = fmt.Sprintf("%s.%s", schema.Name, name)
	}

	return name
}

func (g *QueryGenerator) insertParam(table *Table) (string, string) {
	columns := []string{}
	values := []string{}

	for _, column := range table.Columns {
		columns = append(columns, column.Name)
		values = append(values, "?")
	}

	return strings.Join(columns, ", "), strings.Join(values, ", ")
}

func (g *QueryGenerator) updateParam(table *Table) (string, string) {
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

func (g *QueryGenerator) pkCondition(table *Table) string {
	conditions := []string{}

	for _, column := range table.Columns {
		if !column.Type.IsPrimaryKey {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s = ?", column.Name))
	}

	return strings.Join(conditions, " AND ")
}

func tables(ignore []string, schema *Schema) []Table {
	tables := []Table{}

	if !sort.StringsAreSorted(ignore) {
		sort.Strings(ignore)
	}

	for _, table := range schema.Tables {
		if index := sort.SearchStrings(ignore, table.Name); index >= 0 && index < len(ignore) {
			continue
		}

		tables = append(tables, table)
	}
	return tables
}
