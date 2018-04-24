package sqlmodel_test

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/prana/fake"
	"github.com/phogolabs/prana/sqlmodel"
	"golang.org/x/tools/imports"
)

var _ = Describe("Generator", func() {
	var (
		generator *sqlmodel.Generator
		builder   *fake.ModelTagBuilder
		schemaDef *sqlmodel.Schema
	)

	BeforeEach(func() {
		schemaDef = &sqlmodel.Schema{
			Name:      "schema",
			IsDefault: true,
			Tables: []sqlmodel.Table{
				{
					Name: "table1",
					Columns: []sqlmodel.Column{
						{
							Name:     "id",
							ScanType: "string",
							Type: sqlmodel.ColumnType{
								Name:          "varchar",
								IsPrimaryKey:  true,
								IsNullable:    true,
								CharMaxLength: 200,
							},
						},
						{
							Name:     "name",
							ScanType: "string",
							Type: sqlmodel.ColumnType{
								Name:          "varchar",
								IsPrimaryKey:  false,
								IsNullable:    false,
								CharMaxLength: 200,
							},
						},
					},
				},
			},
		}

		builder = &fake.ModelTagBuilder{}
		builder.BuildReturns("`db`")
		generator = &sqlmodel.Generator{
			TagBuilder: builder,
			Config: &sqlmodel.GeneratorConfig{
				KeepSchema: true,
				InlcudeDoc: false,
			},
		}
	})

	ItGeneratesTheModelSuccessfully := func(table string) {
		It("generates the schema successfully", func() {
			source := &bytes.Buffer{}
			fmt.Fprintln(source, "package model")
			fmt.Fprintln(source)
			fmt.Fprintf(source, "type %s struct {", table)
			fmt.Fprintln(source, "        ID string `db`")
			fmt.Fprintln(source, "        Name string `db`")
			fmt.Fprintln(source, "}")

			data, err := imports.Process("model", source.Bytes(), nil)
			Expect(err).To(BeNil())

			data, err = format.Source(data)
			Expect(err).To(BeNil())

			reader, err := generator.GenerateModel("model", schemaDef)
			Expect(err).To(BeNil())

			Expect(builder.BuildCallCount()).To(Equal(2))
			Expect(builder.BuildArgsForCall(0)).To(Equal(&schemaDef.Tables[0].Columns[0]))
			Expect(builder.BuildArgsForCall(1)).To(Equal(&schemaDef.Tables[0].Columns[1]))

			generated, err := ioutil.ReadAll(reader)
			Expect(err).To(BeNil())
			Expect(string(generated)).To(Equal(string(data)))
		})
	}

	ItGeneratesTheScriptSuccessfully := func(table string) {
		It("generates the SQL script successfully", func() {
			w := &bytes.Buffer{}
			t := strings.Replace(table, ".", "-", -1)
			fmt.Fprintf(w, "-- name: select-all-%s\n", t)
			fmt.Fprintf(w, "SELECT * FROM %s\n\n", table)
			fmt.Fprintf(w, "-- name: select-%s\n", t)
			fmt.Fprintf(w, "SELECT * FROM %s\n", table)
			fmt.Fprint(w, "WHERE id = ?\n\n")
			fmt.Fprintf(w, "-- name: insert-%s\n", t)
			fmt.Fprintf(w, "INSERT INTO %s (id, name)\n", table)
			fmt.Fprint(w, "VALUES (?, ?)\n\n")
			fmt.Fprintf(w, "-- name: update-%s\n", t)
			fmt.Fprintf(w, "UPDATE %s\n", table)
			fmt.Fprint(w, "SET name = ?\n")
			fmt.Fprint(w, "WHERE id = ?\n\n")
			fmt.Fprintf(w, "-- name: delete-%s\n", t)
			fmt.Fprintf(w, "DELETE FROM %s\n", table)
			fmt.Fprint(w, "WHERE id = ?")

			reader, err := generator.GenerateSQLScript(schemaDef)
			Expect(err).To(BeNil())

			generated, err := ioutil.ReadAll(reader)
			Expect(err).To(BeNil())

			Expect(string(generated)).To(Equal(w.String()))
		})
	}

	ItGeneratesTheModelSuccessfully("Table1")
	ItGeneratesTheScriptSuccessfully("schema.table1")

	Context("when KeepSchema is disabled", func() {
		BeforeEach(func() {
			generator.Config.KeepSchema = false
		})

		ItGeneratesTheModelSuccessfully("Table1")
		ItGeneratesTheScriptSuccessfully("table1")

		Context("when the schema is not default", func() {
			BeforeEach(func() {
				schemaDef.IsDefault = false
			})

			ItGeneratesTheModelSuccessfully("ModelTable1")
			ItGeneratesTheScriptSuccessfully("schema.table1")
		})
	})

	Context("when the table is ignored", func() {
		BeforeEach(func() {
			generator.Config.IgnoreTables = []string{"table1"}
		})

		It("generates the schema successfully", func() {
			reader, err := generator.GenerateModel("model", schemaDef)
			Expect(err).To(BeNil())

			data, err := ioutil.ReadAll(reader)
			Expect(err).To(BeNil())
			Expect(data).To(BeEmpty())
		})

		It("generates the commands successfully", func() {
			reader, err := generator.GenerateSQLScript(schemaDef)
			Expect(err).To(BeNil())

			data, err := ioutil.ReadAll(reader)
			Expect(err).To(BeNil())
			Expect(data).To(BeEmpty())
		})
	})

	Context("when including documentation is disabled", func() {
		BeforeEach(func() {
			generator.Config.InlcudeDoc = true
		})

		It("generates the schema successfully", func() {
			reader, err := generator.GenerateModel("model", schemaDef)
			Expect(err).To(BeNil())

			data, err := ioutil.ReadAll(reader)
			Expect(err).To(BeNil())

			source := string(data)
			Expect(source).To(ContainSubstring("// Table1 represents a data base table 'table1'"))
			Expect(source).To(ContainSubstring("// ID represents a database column 'id' of type 'VARCHAR(200) PRIMARY KEY NULL'"))
		})
	})

	Context("when no tables are provided", func() {
		BeforeEach(func() {
			schemaDef.Tables = []sqlmodel.Table{}
		})

		It("generates the schema successfully", func() {
			reader, err := generator.GenerateModel("model", schemaDef)
			Expect(err).To(BeNil())

			data, err := ioutil.ReadAll(reader)
			Expect(err).To(BeNil())
			Expect(data).To(BeEmpty())
		})

		It("generates the script successfully", func() {
			reader, err := generator.GenerateSQLScript(schemaDef)
			Expect(err).To(BeNil())

			data, err := ioutil.ReadAll(reader)
			Expect(err).To(BeNil())
			Expect(data).To(BeEmpty())
		})
	})

	Context("when the package name is not provided", func() {
		It("returns an error", func() {
			reader, err := generator.GenerateModel("", schemaDef)
			Expect(reader).To(BeNil())
			Expect(err).To(MatchError("model:2:1: expected 'IDENT', found 'type'"))
		})
	})
})
