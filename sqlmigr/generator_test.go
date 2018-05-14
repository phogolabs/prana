package sqlmigr_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/parcello"
	"github.com/phogolabs/prana/fake"
	"github.com/phogolabs/prana/sqlmigr"
)

var _ = Describe("Generator", func() {
	var (
		generator *sqlmigr.Generator
		item      *sqlmigr.Migration
		dir       string
	)

	BeforeEach(func() {
		var err error

		dir, err = ioutil.TempDir("", "prana_generator")
		Expect(err).To(BeNil())

		dir = filepath.Join(dir, "sqlmigr")

		generator = &sqlmigr.Generator{
			FileSystem: parcello.Dir(dir),
		}

		item = &sqlmigr.Migration{
			ID:          "20160102150",
			Description: "schema",
			Drivers:     []string{"sql"},
		}
	})

	Describe("Create", func() {
		It("creates a migration successfully", func() {
			err := generator.Create(item)
			Expect(err).To(BeNil())

			path := filepath.Join(dir, item.Filenames()[0])
			Expect(path).To(BeARegularFile())
			Expect(dir).To(BeADirectory())

			data, err := ioutil.ReadFile(path)
			Expect(err).To(BeNil())

			script := string(data)
			Expect(script).To(ContainSubstring("-- name: up"))
			Expect(script).To(ContainSubstring("-- name: down"))
		})

		Context("when the dir is the root dir", func() {
			It("returns an error", func() {
				generator.FileSystem = parcello.Dir("/")
				err := generator.Create(item)
				Expect(err.Error()).To(Equal("open /20160102150_schema.sql: permission denied"))
			})
		})
	})

	Describe("Write", func() {
		It("writes a sqlmigr successfully", func() {
			content := &sqlmigr.Content{
				UpCommand:   bytes.NewBufferString("upgrade"),
				DownCommand: bytes.NewBufferString("rollback"),
			}

			Expect(generator.Write(item, content)).To(Succeed())
			Expect(dir).To(BeADirectory())

			path := filepath.Join(dir, item.Filenames()[0])
			Expect(path).To(BeARegularFile())

			data, err := ioutil.ReadFile(path)
			Expect(err).To(BeNil())

			script := string(data)
			Expect(script).To(ContainSubstring("-- name: up"))
			Expect(script).To(ContainSubstring("upgrade"))
			Expect(script).To(ContainSubstring("-- name: down"))
			Expect(script).To(ContainSubstring("rollback"))
		})

		Context("when writing to the fails fails", func() {
			It("returns an error", func() {
				content := &sqlmigr.Content{
					UpCommand:   bytes.NewBufferString("commit"),
					DownCommand: bytes.NewBufferString("rollback"),
				}

				writer := &fake.File{}
				writer.WriteReturns(1, nil)

				fileSystem := &fake.FileSystem{}
				fileSystem.OpenFileReturns(writer, nil)

				generator.FileSystem = fileSystem

				Expect(generator.Write(item, content)).To(MatchError("short write"))
			})
		})

		Context("when the dir is not valid", func() {
			It("returns an error", func() {
				content := &sqlmigr.Content{
					UpCommand:   bytes.NewBufferString("upgrade"),
					DownCommand: bytes.NewBufferString("rollback"),
				}
				generator.FileSystem = parcello.Dir("/")
				Expect(generator.Write(item, content)).To(MatchError("open /20160102150_schema.sql: permission denied"))
			})
		})

		Context("when the up step cannot be created", func() {
			It("returns an error", func() {
				reader := &fake.File{}
				content := &sqlmigr.Content{
					UpCommand:   reader,
					DownCommand: bytes.NewBufferString("rollback"),
				}
				reader.ReadReturns(0, fmt.Errorf("Oh no!"))
				Expect(generator.Write(item, content)).To(MatchError("Oh no!"))
			})
		})

		Context("when the up step cannot be created", func() {
			It("returns an error", func() {
				reader := &fake.File{}
				content := &sqlmigr.Content{
					UpCommand:   bytes.NewBufferString("upgrade"),
					DownCommand: reader,
				}
				reader.ReadReturns(0, fmt.Errorf("Oh no!"))
				Expect(generator.Write(item, content)).To(MatchError("Oh no!"))
			})
		})
	})
})
