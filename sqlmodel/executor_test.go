package sqlmodel_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/phogolabs/prana/fake"
	"github.com/phogolabs/prana/sqlmodel"
	"github.com/phogolabs/prana/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Executor", func() {
	var (
		directory string
		executor  *sqlmodel.Executor
		spec      *sqlmodel.Spec
		provider  *fake.SchemaProvider
		composer  *fake.ModelGenerator
		schemaDef *sqlmodel.Schema
	)

	BeforeEach(func() {
		schemaDef = &sqlmodel.Schema{
			Name:      "public",
			IsDefault: true,
			Tables: []sqlmodel.Table{
				{
					Name: "table1",
					Columns: []sqlmodel.Column{
						{
							Name:     "ID",
							ScanType: "string",
						},
					},
				},
			},
		}

		var err error
		directory, err = ioutil.TempDir("", "prana")
		Expect(err).To(BeNil())

		spec = &sqlmodel.Spec{
			Schema:     "public",
			Template:   "model",
			Tables:     []string{"table1"},
			Filename:   "schema.go",
			FileSystem: storage.New(directory),
		}

		provider = &fake.SchemaProvider{}
		provider.TablesReturns([]string{"table1"}, nil)
		provider.SchemaReturns(schemaDef, nil)

		composer = &fake.ModelGenerator{}
		executor = &sqlmodel.Executor{
			Provider:  provider,
			Generator: composer,
		}
	})

	Describe("Write", func() {
		BeforeEach(func() {
			composer.GenerateStub = func(ctx *sqlmodel.GeneratorContext) error {
				ctx.Writer.Write([]byte("source"))
				return nil
			}
		})

		It("writes the generated source successfully", func() {
			writer := &bytes.Buffer{}
			Expect(executor.Write(writer, spec)).To(Succeed())
			Expect(writer.String()).To(Equal("source"))

			Expect(provider.TablesCallCount()).To(BeZero())
			Expect(provider.SchemaCallCount()).To(Equal(1))

			schemaName, tables := provider.SchemaArgsForCall(0)
			Expect(schemaName).To(Equal("public"))
			Expect(tables).To(ContainElement("table1"))

			Expect(composer.GenerateCallCount()).To(Equal(1))
			ctx := composer.GenerateArgsForCall(0)

			Expect(ctx.Schema).To(Equal(schemaDef))
		})

		Context("when the schema is not default", func() {
			BeforeEach(func() {
				schemaDef.IsDefault = false
			})

			It("uses the schema name as package name", func() {
				Expect(executor.Write(ioutil.Discard, spec)).To(Succeed())
				Expect(composer.GenerateCallCount()).To(Equal(1))
				ctx := composer.GenerateArgsForCall(0)

				Expect(ctx.Schema).To(Equal(schemaDef))
			})
		})

		Context("when the tables are not provided", func() {
			BeforeEach(func() {
				spec.Tables = []string{}
			})

			It("writes the generated source successfully", func() {
				writer := &bytes.Buffer{}

				Expect(executor.Write(writer, spec)).To(Succeed())
				Expect(writer.String()).To(Equal("source"))

				Expect(provider.TablesCallCount()).To(Equal(1))
				Expect(provider.TablesArgsForCall(0)).To(Equal("public"))

				Expect(provider.SchemaCallCount()).To(Equal(1))

				schemaName, tables := provider.SchemaArgsForCall(0)
				Expect(schemaName).To(Equal("public"))
				Expect(tables).To(ContainElement("table1"))

				Expect(composer.GenerateCallCount()).To(Equal(1))
				ctx := composer.GenerateArgsForCall(0)

				Expect(ctx.Schema).To(Equal(schemaDef))
			})

			Context("when getting the schema tables fails", func() {
				BeforeEach(func() {
					provider.TablesReturns([]string{}, fmt.Errorf("Oh no!"))
				})

				It("returns the error", func() {
					writer := &bytes.Buffer{}
					Expect(executor.Write(writer, spec)).To(MatchError("Oh no!"))
					Expect(writer.Bytes()).To(BeEmpty())
				})
			})
		})

		Context("when the composer fails", func() {
			BeforeEach(func() {
				composer.GenerateReturns(fmt.Errorf("Oh no!"))
			})

			It("returns the error", func() {
				Expect(executor.Write(ioutil.Discard, spec)).To(MatchError("Oh no!"))
			})
		})
	})

	Describe("Create", func() {
		BeforeEach(func() {
			composer.GenerateStub = func(ctx *sqlmodel.GeneratorContext) error {
				ctx.Writer.Write([]byte("source"))
				return nil
			}
		})

		ItCreatesTheSchemaInRootPkg := func(filename, pkg string) {
			It("creates a package with generated source successfully", func() {
				path, err := executor.Create(spec)
				Expect(err).To(Succeed())
				Expect(path).To(Equal(filename))

				dir := fmt.Sprintf("%v", directory)
				Expect(dir).To(BeADirectory())
				Expect(filepath.Join(dir, path)).To(BeARegularFile())

				Expect(provider.TablesCallCount()).To(BeZero())
				Expect(provider.SchemaCallCount()).To(Equal(1))

				schemaName, tables := provider.SchemaArgsForCall(0)
				Expect(schemaName).To(Equal("public"))
				Expect(tables).To(ContainElement("table1"))

				Expect(composer.GenerateCallCount()).To(Equal(1))
				ctx := composer.GenerateArgsForCall(0)

				Expect(ctx.Schema).To(Equal(schemaDef))
			})
		}

		ItCreatesTheSchemaInRootPkg("schema.go", "entity")

		Context("when the schema is not default", func() {
			BeforeEach(func() {
				schemaDef.IsDefault = false
			})

			ItCreatesTheSchemaInRootPkg("public.go", "entity")
		})

		Context("when the KeepSchema is false", func() {
			ItCreatesTheSchemaInRootPkg("schema.go", "entity")

			Context("when the schema is not default", func() {
				BeforeEach(func() {
					schemaDef.IsDefault = false
				})

				ItCreatesTheSchemaInRootPkg("public.go", "entity")
			})
		})

		Context("when the tables are not provided", func() {
			BeforeEach(func() {
				spec.Tables = []string{}
			})

			It("creates a package with generated source successfully", func() {
				path, err := executor.Create(spec)
				Expect(err).To(Succeed())
				Expect(path).To(Equal("schema.go"))

				Expect(provider.TablesCallCount()).To(Equal(1))
				Expect(provider.TablesArgsForCall(0)).To(Equal("public"))
				Expect(provider.SchemaCallCount()).To(Equal(1))

				schemaName, tables := provider.SchemaArgsForCall(0)
				Expect(schemaName).To(Equal("public"))
				Expect(tables).To(ContainElement("table1"))

				Expect(composer.GenerateCallCount()).To(Equal(1))
				ctx := composer.GenerateArgsForCall(0)

				Expect(ctx.Schema).To(Equal(schemaDef))
			})

			Context("when the provider fails to get table names", func() {
				BeforeEach(func() {
					provider.TablesReturns([]string{}, fmt.Errorf("Oh no!"))
				})

				It("returns the error", func() {
					path, err := executor.Create(spec)
					Expect(err).To(MatchError("Oh no!"))
					Expect(path).To(BeEmpty())
				})
			})
		})

		Context("when the spec schema is not default", func() {
			BeforeEach(func() {
				schemaDef.IsDefault = false
			})

			It("creates a package with generated source successfully", func() {
				path, err := executor.Create(spec)
				Expect(err).To(Succeed())
				Expect(path).To(Equal("public.go"))
			})
		})

		Context("when getting the schame fails", func() {
			BeforeEach(func() {
				provider.SchemaReturns(nil, fmt.Errorf("Oh no!"))
			})

			It("returns the error", func() {
				path, err := executor.Create(spec)
				Expect(err).To(MatchError("Oh no!"))
				Expect(path).To(BeEmpty())
			})
		})

		Context("when the reader has empty content", func() {
			BeforeEach(func() {
				composer.GenerateStub = func(ctx *sqlmodel.GeneratorContext) error {
					return nil
				}
			})

			It("creates a package with generated source successfully", func() {
				path, err := executor.Create(spec)
				Expect(err).To(Succeed())
				Expect(path).To(BeEmpty())
			})
		})

		Context("when the composer fails", func() {
			BeforeEach(func() {
				composer.GenerateReturns(fmt.Errorf("Oh no!"))
			})

			It("returns the error", func() {
				path, err := executor.Create(spec)
				Expect(err).To(MatchError("Oh no!"))
				Expect(path).To(BeEmpty())
			})
		})
	})
})
