package sqlmodel_test

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/prana/fake"
	"github.com/phogolabs/prana/sqlmodel"
	"golang.org/x/tools/imports"
)

var _ = Describe("ModelGenerator", func() {
	var (
		generator *sqlmodel.ModelGenerator
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

		generator = &sqlmodel.ModelGenerator{
			TagBuilder: builder,
			Config: &sqlmodel.ModelGeneratorConfig{
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

			reader := &bytes.Buffer{}

			ctx := &sqlmodel.GeneratorContext{
				Writer:  reader,
				Package: "model",
				Schema:  schemaDef,
			}

			Expect(generator.Generate(ctx)).To(Succeed())
			Expect(builder.BuildCallCount()).To(Equal(2))
			Expect(builder.BuildArgsForCall(0)).To(Equal(&schemaDef.Tables[0].Columns[0]))
			Expect(builder.BuildArgsForCall(1)).To(Equal(&schemaDef.Tables[0].Columns[1]))
			Expect(reader.String()).To(Equal(string(data)))
		})
	}

	ItGeneratesTheModelSuccessfully("Table1")

	Context("when KeepSchema is disabled", func() {
		BeforeEach(func() {
			generator.Config.KeepSchema = false
		})

		ItGeneratesTheModelSuccessfully("Table1")

		Context("when the schema is not default", func() {
			BeforeEach(func() {
				schemaDef.IsDefault = false
			})

			ItGeneratesTheModelSuccessfully("ModelTable1")
		})
	})

	Context("when the table is ignored", func() {
		BeforeEach(func() {
			generator.Config.IgnoreTables = []string{"table1"}
		})

		It("generates the schema successfully", func() {
			reader := &bytes.Buffer{}
			ctx := &sqlmodel.GeneratorContext{
				Writer:  reader,
				Package: "model",
				Schema:  schemaDef,
			}

			Expect(generator.Generate(ctx)).To(Succeed())
			Expect(reader.String()).To(BeEmpty())
		})
	})

	Context("when including documentation is disabled", func() {
		BeforeEach(func() {
			generator.Config.InlcudeDoc = true
		})

		It("generates the schema successfully", func() {
			reader := &bytes.Buffer{}
			ctx := &sqlmodel.GeneratorContext{
				Writer:  reader,
				Package: "model",
				Schema:  schemaDef,
			}

			Expect(generator.Generate(ctx)).To(Succeed())
			Expect(reader.String()).To(ContainSubstring("// Table1 represents a data base table 'table1'"))
			Expect(reader.String()).To(ContainSubstring("// ID represents a database column 'id' of type 'VARCHAR(200) PRIMARY KEY NULL'"))
		})
	})

	Context("when no tables are provided", func() {
		BeforeEach(func() {
			schemaDef.Tables = []sqlmodel.Table{}
		})

		It("generates the schema successfully", func() {
			reader := &bytes.Buffer{}
			ctx := &sqlmodel.GeneratorContext{
				Writer:  reader,
				Package: "model",
				Schema:  schemaDef,
			}

			Expect(generator.Generate(ctx)).To(Succeed())
			Expect(reader.String()).To(BeEmpty())
		})
	})

	Context("when the package name is not provided", func() {
		It("returns an error", func() {
			reader := &bytes.Buffer{}
			ctx := &sqlmodel.GeneratorContext{
				Writer: reader,
				Schema: schemaDef,
			}
			err := generator.Generate(ctx)
			Expect(err).To(MatchError("model:2:1: expected 'IDENT', found 'type'"))
		})
	})
})

var _ = Describe("QueryGenerator", func() {
	var (
		generator *sqlmodel.QueryGenerator
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

		generator = &sqlmodel.QueryGenerator{
			Config: &sqlmodel.QueryGeneratorConfig{
				InlcudeDoc: false,
			},
		}
	})

	ItGeneratesTheScriptSuccessfully := func(table string) {
		It("generates the SQL script successfully", func() {
			w := &bytes.Buffer{}
			t := strings.Replace(table, ".", "-", -1)
			fmt.Fprintf(w, "-- name: select-all-%s\n", t)
			fmt.Fprintf(w, "SELECT * FROM %s;\n\n", table)
			fmt.Fprintf(w, "-- name: select-%s\n", t)
			fmt.Fprintf(w, "SELECT * FROM %s\n", table)
			fmt.Fprint(w, "WHERE id = ?;\n\n")
			fmt.Fprintf(w, "-- name: insert-%s\n", t)
			fmt.Fprintf(w, "INSERT INTO %s (id, name)\n", table)
			fmt.Fprint(w, "VALUES (?, ?);\n\n")
			fmt.Fprintf(w, "-- name: update-%s\n", t)
			fmt.Fprintf(w, "UPDATE %s\n", table)
			fmt.Fprint(w, "SET name = ?\n")
			fmt.Fprint(w, "WHERE id = ?;\n\n")
			fmt.Fprintf(w, "-- name: delete-%s\n", t)
			fmt.Fprintf(w, "DELETE FROM %s\n", table)
			fmt.Fprint(w, "WHERE id = ?;")

			reader := &bytes.Buffer{}
			ctx := &sqlmodel.GeneratorContext{
				Writer:  reader,
				Package: "model",
				Schema:  schemaDef,
			}

			Expect(generator.Generate(ctx)).To(Succeed())
			Expect(reader.String()).To(Equal(w.String()))
		})
	}

	ItGeneratesTheScriptSuccessfully("table1")

	Context("when the schema is not default", func() {
		BeforeEach(func() {
			schemaDef.IsDefault = false
		})

		ItGeneratesTheScriptSuccessfully("schema.table1")
	})

	Context("when named parameters are used", func() {
		BeforeEach(func() {
			generator.Config.UseNamedParams = true
		})

		It("generates the SQL script successfully", func() {
			table := "table1"
			w := &bytes.Buffer{}
			t := strings.Replace(table, ".", "-", -1)
			fmt.Fprintf(w, "-- name: select-all-%s\n", t)
			fmt.Fprintf(w, "SELECT * FROM %s;\n\n", table)
			fmt.Fprintf(w, "-- name: select-%s\n", t)
			fmt.Fprintf(w, "SELECT * FROM %s\n", table)
			fmt.Fprint(w, "WHERE id = :id;\n\n")
			fmt.Fprintf(w, "-- name: insert-%s\n", t)
			fmt.Fprintf(w, "INSERT INTO %s (id, name)\n", table)
			fmt.Fprint(w, "VALUES (:id, :name);\n\n")
			fmt.Fprintf(w, "-- name: update-%s\n", t)
			fmt.Fprintf(w, "UPDATE %s\n", table)
			fmt.Fprint(w, "SET name = :name\n")
			fmt.Fprint(w, "WHERE id = :id;\n\n")
			fmt.Fprintf(w, "-- name: delete-%s\n", t)
			fmt.Fprintf(w, "DELETE FROM %s\n", table)
			fmt.Fprint(w, "WHERE id = :id;")

			reader := &bytes.Buffer{}
			ctx := &sqlmodel.GeneratorContext{
				Writer:  reader,
				Package: "model",
				Schema:  schemaDef,
			}

			Expect(generator.Generate(ctx)).To(Succeed())
			Expect(reader.String()).To(Equal(w.String()))
		})
	})

	Context("when the table is ignored", func() {
		BeforeEach(func() {
			generator.Config.IgnoreTables = []string{"table1"}
		})

		It("generates the commands successfully", func() {
			reader := &bytes.Buffer{}
			ctx := &sqlmodel.GeneratorContext{
				Writer:  reader,
				Package: "model",
				Schema:  schemaDef,
			}

			Expect(generator.Generate(ctx)).To(Succeed())
			Expect(reader.String()).To(BeEmpty())
		})
	})

	Context("when including documentation is disabled", func() {
		BeforeEach(func() {
			generator.Config.InlcudeDoc = true
		})

		It("generates the schema successfully", func() {
			reader := &bytes.Buffer{}
			ctx := &sqlmodel.GeneratorContext{
				Writer:  reader,
				Package: "model",
				Schema:  schemaDef,
			}

			Expect(generator.Generate(ctx)).To(Succeed())
			Expect(reader.String()).To(ContainSubstring("-- Auto-generated"))
		})
	})

	Context("when no tables are provided", func() {
		BeforeEach(func() {
			schemaDef.Tables = []sqlmodel.Table{}
		})

		It("generates the script successfully", func() {
			reader := &bytes.Buffer{}
			ctx := &sqlmodel.GeneratorContext{
				Writer:  reader,
				Package: "model",
				Schema:  schemaDef,
			}

			Expect(generator.Generate(ctx)).To(Succeed())
			Expect(reader.String()).To(BeEmpty())
		})
	})
})
