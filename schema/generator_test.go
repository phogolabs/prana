package schema_test

import (
	"bytes"
	"fmt"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/gom/schema"
	"golang.org/x/tools/imports"
)

var _ = Describe("Generator", func() {
	var (
		generator *schema.Generator
		source    *bytes.Buffer
		schemaDef *schema.Schema
	)

	BeforeEach(func() {
		schemaDef = &schema.Schema{
			Name: "schema",
			Tables: []schema.Table{
				schema.Table{
					Name: "table1",
					Columns: []schema.Column{
						schema.Column{
							Name:     "ID",
							ScanType: "string",
						},
					},
				},
			},
		}

		generator = &schema.Generator{
			Config: &schema.GeneratorConfig{
				InlcudeDoc: false,
			},
		}

		source = &bytes.Buffer{}
		fmt.Fprintln(source, "package model")
		fmt.Fprintln(source)
		fmt.Fprintln(source, "type Table1 struct {")
		fmt.Fprintln(source, "  ID string `db:\"ID\" json:\"ID\" validate:\"required\"`")
		fmt.Fprintln(source, "}")

		data, err := imports.Process("model", source.Bytes(), nil)
		Expect(err).To(BeNil())

		source.Reset()
		source.Write(data)
	})

	It("generates the schema successfully", func() {
		reader, err := generator.Compose("model", schemaDef)
		Expect(err).To(BeNil())

		data, err := ioutil.ReadAll(reader)
		Expect(err).To(BeNil())
		Expect(data).To(Equal(source.Bytes()))
	})

	Context("when including documentation is enabled", func() {
		BeforeEach(func() {
			generator.Config.InlcudeDoc = true
		})

		It("generates the schema successfully", func() {
			reader, err := generator.Compose("model", schemaDef)
			Expect(err).To(BeNil())

			data, err := ioutil.ReadAll(reader)
			Expect(err).To(BeNil())

			source := string(data)
			Expect(source).To(ContainSubstring("// ID represents a database column 'ID' of type ' NOT NULL'"))
			Expect(source).To(ContainSubstring("// Table1 represents a data base table 'table1'"))
		})
	})

	Context("when no tables are provided", func() {
		BeforeEach(func() {
			schemaDef.Tables = []schema.Table{}
		})

		It("generates the schema successfully", func() {
			reader, err := generator.Compose("model", schemaDef)
			Expect(err).To(BeNil())

			data, err := ioutil.ReadAll(reader)
			Expect(err).To(BeNil())
			Expect(data).To(BeEmpty())
		})
	})
})
