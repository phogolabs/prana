package model_test

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/oak/fake"
	"github.com/phogolabs/oak/model"
	"golang.org/x/tools/imports"
)

var _ = Describe("Generator", func() {
	var (
		generator *model.Generator
		builder   *fake.ModelTagBuilder
		schemaDef *model.Schema
	)

	BeforeEach(func() {
		schemaDef = &model.Schema{
			Name:      "schema",
			IsDefault: true,
			Tables: []model.Table{
				{
					Name: "table1",
					Columns: []model.Column{
						{
							Name:     "id",
							ScanType: "string",
							Type: model.ColumnType{
								Name:          "varchar",
								IsPrimaryKey:  true,
								IsNullable:    true,
								CharMaxLength: 200,
							},
						},
						{
							Name:     "name",
							ScanType: "string",
							Type: model.ColumnType{
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
		generator = &model.Generator{
			TagBuilder: builder,
			Config: &model.GeneratorConfig{
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
			fmt.Fprintln(source, "        Id string `db`")
			fmt.Fprintln(source, "        Name string `db`")
			fmt.Fprintln(source, "}")

			data, err := imports.Process("model", source.Bytes(), nil)
			Expect(err).To(BeNil())

			data, err = format.Source(data)
			Expect(err).To(BeNil())

			reader, err := generator.Compose("model", schemaDef)
			Expect(err).To(BeNil())

			Expect(builder.BuildCallCount()).To(Equal(2))
			Expect(builder.BuildArgsForCall(0)).To(Equal(&schemaDef.Tables[0].Columns[0]))
			Expect(builder.BuildArgsForCall(1)).To(Equal(&schemaDef.Tables[0].Columns[1]))

			generated, err := ioutil.ReadAll(reader)
			Expect(err).To(BeNil())
			Expect(string(generated)).To(Equal(string(data)))
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
			reader, err := generator.Compose("model", schemaDef)
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
			reader, err := generator.Compose("model", schemaDef)
			Expect(err).To(BeNil())

			data, err := ioutil.ReadAll(reader)
			Expect(err).To(BeNil())

			source := string(data)
			Expect(source).To(ContainSubstring("// Table1 represents a data base table 'table1'"))
			Expect(source).To(ContainSubstring("// Id represents a database column 'id' of type 'VARCHAR(200) PRIMARY KEY NULL'"))
		})
	})

	Context("when the tables are ignored", func() {
		BeforeEach(func() {
			generator.Config.IgnoreTables = []string{"table1", "atab"}
		})

		It("generates the schema successfully", func() {
			reader, err := generator.Compose("model", schemaDef)
			Expect(err).To(BeNil())

			data, err := ioutil.ReadAll(reader)
			Expect(err).To(BeNil())
			Expect(data).To(BeEmpty())
		})
	})

	Context("when no tables are provided", func() {
		BeforeEach(func() {
			schemaDef.Tables = []model.Table{}
		})

		It("generates the schema successfully", func() {
			reader, err := generator.Compose("model", schemaDef)
			Expect(err).To(BeNil())

			data, err := ioutil.ReadAll(reader)
			Expect(err).To(BeNil())
			Expect(data).To(BeEmpty())
		})
	})

	Context("when the package name is not provided", func() {
		It("returns an error", func() {
			reader, err := generator.Compose("", schemaDef)
			Expect(reader).To(BeNil())
			Expect(err).To(MatchError("model:2:1: expected 'IDENT', found 'type'"))
		})
	})
})
