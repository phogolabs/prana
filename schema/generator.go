package schema

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/inflection"
	"golang.org/x/tools/imports"
)

var _ Composer = &Generator{}

// GeneratorConfig controls how the code generation happens
type GeneratorConfig struct {
	// InlcudeDoc determines whether to include documentation
	InlcudeDoc bool
	// IgnoreTables ecludes the those tables from generation
	IgnoreTables []string
}

// Generator generates Golang structs from database schema
type Generator struct {
	// Config controls how the code generation happens
	Config *GeneratorConfig
}

// Compose generates the golang structs from database schema
func (g *Generator) Compose(pkg string, schema *Schema) (io.Reader, error) {
	buffer := &bytes.Buffer{}

	if len(schema.Tables) == 0 {
		return buffer, nil
	}

	ignore := g.ignore()
	processed := 0

	if g.Config.InlcudeDoc {
		fmt.Fprintf(buffer, "// Package contains an object model of database schema '%s'", schema.Name)
		fmt.Fprintln(buffer)
		fmt.Fprintln(buffer, "// Auto-generated at", time.Now().Format(time.UnixDate))
	}

	fmt.Fprintf(buffer, "package ")
	fmt.Fprintf(buffer, pkg)
	fmt.Fprintln(buffer)

	for _, table := range schema.Tables {
		if index := sort.SearchStrings(ignore, table.Name); index >= 0 && index < len(ignore) {
			continue
		}

		processed = processed + 1
		columns := table.Columns
		length := len(columns)
		typeName := g.tableName(&table)

		if g.Config.InlcudeDoc {
			fmt.Fprintln(buffer)
			fmt.Fprintf(buffer, "// %s represents a data base table '%s'", typeName, table.Name)
			fmt.Fprintln(buffer)
		}

		fmt.Fprintf(buffer, "type %v struct {", typeName)
		fmt.Fprintln(buffer)

		for index, column := range columns {
			fieldName := g.fieldName(&column)
			fieldType := g.fieldType(&column)
			fieldTag := g.fieldTag(&column)

			if g.Config.InlcudeDoc {
				if index > 0 {
					fmt.Fprintln(buffer)
				}
				fmt.Fprintf(buffer, "// %s represents a database column '%s' of type '%s'", fieldName, column.Name, column.Type)
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

	if processed == 0 {
		buffer.Reset()
	} else if err := g.format(buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}

func (g *Generator) ignore() []string {
	ignore := g.Config.IgnoreTables

	if !sort.StringsAreSorted(ignore) {
		sort.Strings(ignore)
	}

	return ignore
}

func (g *Generator) tableName(table *Table) string {
	name := g.sanitize(table.Name)
	name = inflection.Singular(name)
	name = strings.Title(name)
	return name
}

func (g *Generator) fieldName(column *Column) string {
	return g.sanitize(column.Name)
}

func (g *Generator) fieldType(column *Column) string {
	return column.ScanType
}

func (g *Generator) fieldTag(column *Column) string {
	tag := func(name, value string) string {
		return fmt.Sprintf(`%s:"%s"`, name, value)
	}

	tags := []string{}
	tags = append(tags, tag("db", column.Name))
	tags = append(tags, tag("json", column.Name))

	if !column.Type.IsNullable {
		tags = append(tags, tag("validate", "required"))
	}

	return fmt.Sprintf("`%s`", strings.Join(tags, " "))
}

func (g *Generator) sanitize(text string) string {
	buffer := &bytes.Buffer{}
	parts := strings.Split(text, "_")

	for _, part := range parts {
		buffer.WriteString(strings.Title(part))
	}

	return buffer.String()
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
