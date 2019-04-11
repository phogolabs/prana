package sqlmodel_test

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/prana/sqlmodel"
	"golang.org/x/tools/imports"
)

var _ = Describe("Codegen", func() {
	var (
		generator *sqlmodel.Codegen
		schemaDef *sqlmodel.Schema
	)

	BeforeEach(func() {
		schemaDef = NewSchema()
		generator = &sqlmodel.Codegen{}
	})

	Describe("Model", func() {
		BeforeEach(func() {
			generator.Format = true
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

				reader := &bytes.Buffer{}

				ctx := &sqlmodel.GeneratorContext{
					Writer:   reader,
					Template: "model",
					Schema:   schemaDef,
				}

				Expect(generator.Generate(ctx)).To(Succeed())
				Expect(reader.String()).To(Equal(string(data)))
			})
		}

		ItGeneratesTheModelSuccessfully("Table1")

		Context("when the schema is not default", func() {
			BeforeEach(func() {
				schemaDef.IsDefault = false
			})

			ItGeneratesTheModelSuccessfully("Table1")
		})

		Context("when no tables are provided", func() {
			BeforeEach(func() {
				schemaDef.Tables = []sqlmodel.Table{}
			})

			It("generates the schema successfully", func() {
				reader := &bytes.Buffer{}
				ctx := &sqlmodel.GeneratorContext{
					Writer:   reader,
					Template: "model",
					Schema:   schemaDef,
				}

				Expect(generator.Generate(ctx)).To(Succeed())
				Expect(reader.String()).To(BeEmpty())
			})
		})

		Context("when the package name is not provided", func() {
			It("returns an error", func() {
				reader := &bytes.Buffer{}
				schemaDef.Model.Package = ""
				ctx := &sqlmodel.GeneratorContext{
					Writer:   reader,
					Template: "model",
					Schema:   schemaDef,
				}
				err := generator.Generate(ctx)
				Expect(err.Error()).To(Equal("model:3:1: expected 'IDENT', found 'type'"))
			})
		})
	})

	Describe("Routine", func() {
		BeforeEach(func() {
			generator.Format = false
		})

		ItGeneratesTheScriptSuccessfully := func(table string) {
			It("generates the SQL script successfully", func() {
				w := &bytes.Buffer{}
				t := strings.Replace(table, ".", "-", -1)
				fmt.Fprintf(w, "-- name: select-all-%s\n", t)
				fmt.Fprintf(w, "SELECT * FROM %s;\n\n", table)
				fmt.Fprintf(w, "-- name: select-%s-by-pk\n", t)
				fmt.Fprintf(w, "SELECT * FROM %s\n", table)
				fmt.Fprint(w, "WHERE id = ?;\n\n")
				fmt.Fprintf(w, "-- name: insert-%s\n", t)
				fmt.Fprintf(w, "INSERT INTO %s (id, name)\n", table)
				fmt.Fprint(w, "VALUES (?, ?);\n\n")
				fmt.Fprintf(w, "-- name: update-%s-by-pk\n", t)
				fmt.Fprintf(w, "UPDATE %s\n", table)
				fmt.Fprint(w, "SET name = ?\n")
				fmt.Fprint(w, "WHERE id = ?;\n\n")
				fmt.Fprintf(w, "-- name: delete-%s-by-pk\n", t)
				fmt.Fprintf(w, "DELETE FROM %s\n", table)
				fmt.Fprint(w, "WHERE id = ?;\n\n")

				reader := &bytes.Buffer{}
				ctx := &sqlmodel.GeneratorContext{
					Writer:   reader,
					Template: "routine",
					Schema:   schemaDef,
				}

				Expect(generator.Generate(ctx)).To(Succeed())
				Expect(reader.String()).To(Equal(w.String()))
			})
		}

		ItGeneratesTheScriptSuccessfully("table1")

		Context("when no tables are provided", func() {
			BeforeEach(func() {
				schemaDef.Tables = []sqlmodel.Table{}
			})

			It("generates the schema successfully", func() {
				reader := &bytes.Buffer{}
				ctx := &sqlmodel.GeneratorContext{
					Writer:   reader,
					Template: "routine",
					Schema:   schemaDef,
				}

				Expect(generator.Generate(ctx)).To(Succeed())
				Expect(reader.String()).To(BeEmpty())
			})
		})

		Context("when the table does not have columns", func() {
			BeforeEach(func() {
				schemaDef.Tables[0].Columns = []sqlmodel.Column{}
			})

			It("generates the schema successfully", func() {
				reader := &bytes.Buffer{}
				ctx := &sqlmodel.GeneratorContext{
					Writer:   reader,
					Template: "routine",
					Schema:   schemaDef,
				}

				Expect(generator.Generate(ctx)).To(Succeed())
				Expect(reader.String()).To(ContainSubstring("select-all-table1"))
			})
		})

		Context("when more than one table are provided", func() {
			BeforeEach(func() {
				schemaDef.Tables = append(schemaDef.Tables,
					sqlmodel.Table{
						Name: "table2",
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
				)
			})

			It("generates the script successfully", func() {
				reader := &bytes.Buffer{}
				ctx := &sqlmodel.GeneratorContext{
					Writer:   reader,
					Template: "routine",
					Schema:   schemaDef,
				}

				Expect(generator.Generate(ctx)).To(Succeed())
				Expect(reader.String()).To(ContainSubstring("table1"))
				Expect(reader.String()).To(ContainSubstring("table2"))
			})
		})
	})

	Describe("Repository", func() {
		BeforeEach(func() {
			generator.Format = true
			generator.Meta = map[string]interface{}{
				"RepositoryPackage": "model",
			}
		})

		ItGeneratesTheRepositorySuccessfully := func(table string) {
			It("generates the repository successfully", func() {
				source, err := ioutil.ReadFile("./fixture/repository.txt")
				Expect(err).To(BeNil())

				data, err := imports.Process("model", source, nil)
				Expect(err).To(BeNil())

				data, err = format.Source(data)
				Expect(err).To(BeNil())

				reader := &bytes.Buffer{}

				ctx := &sqlmodel.GeneratorContext{
					Writer:   reader,
					Template: "repository",
					Schema:   schemaDef,
				}

				Expect(generator.Generate(ctx)).To(Succeed())
				Expect(reader.String()).To(Equal(string(data)))
			})
		}

		ItGeneratesTheRepositorySuccessfully("Table1")

		Context("when the schema is not default", func() {
			BeforeEach(func() {
				schemaDef.IsDefault = false
			})

			ItGeneratesTheRepositorySuccessfully("Table1")
		})

		Context("when no tables are provided", func() {
			BeforeEach(func() {
				schemaDef.Tables = []sqlmodel.Table{}
			})

			It("generates the schema successfully", func() {
				reader := &bytes.Buffer{}
				ctx := &sqlmodel.GeneratorContext{
					Writer:   reader,
					Template: "repository",
					Schema:   schemaDef,
				}

				Expect(generator.Generate(ctx)).To(Succeed())
				Expect(reader.String()).To(BeEmpty())
			})
		})

		Context("when the package name is not provided", func() {
			It("returns an error", func() {
				generator.Meta = map[string]interface{}{}
				reader := &bytes.Buffer{}
				schemaDef.Model.Package = ""
				ctx := &sqlmodel.GeneratorContext{
					Writer:   reader,
					Template: "repository",
					Schema:   schemaDef,
				}
				err := generator.Generate(ctx)
				Expect(err.Error()).To(Equal("repository:3:1: expected 'IDENT', found 'type'"))
			})
		})
	})
})
