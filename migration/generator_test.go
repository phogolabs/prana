package migration_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/oak/fake"
	"github.com/phogolabs/oak/migration"
	"github.com/phogolabs/parcel"
)

var _ = Describe("Generator", func() {
	var (
		generator *migration.Generator
		item      *migration.Item
		dir       string
	)

	BeforeEach(func() {
		var err error

		dir, err = ioutil.TempDir("", "oak_generator")
		Expect(err).To(BeNil())

		dir = filepath.Join(dir, "migration")

		generator = &migration.Generator{
			FileSystem: parcel.Dir(dir),
		}

		item = &migration.Item{
			ID:          "20160102150",
			Description: "schema",
		}
	})

	Describe("Create", func() {
		It("creates a migration successfully", func() {
			err := generator.Create(item)
			Expect(err).To(BeNil())

			path := filepath.Join(dir, item.Filename())
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
				generator.FileSystem = parcel.Dir("/")
				err := generator.Create(item)
				Expect(err.Error()).To(Equal("open /20160102150_schema.sql: permission denied"))
			})
		})
	})

	Describe("Write", func() {
		It("writes a migration successfully", func() {
			content := &migration.Content{
				UpCommand:   bytes.NewBufferString("upgrade"),
				DownCommand: bytes.NewBufferString("rollback"),
			}

			Expect(generator.Write(item, content)).To(Succeed())
			Expect(dir).To(BeADirectory())

			path := filepath.Join(dir, item.Filename())
			Expect(path).To(BeARegularFile())

			data, err := ioutil.ReadFile(path)
			Expect(err).To(BeNil())

			script := string(data)
			Expect(script).To(ContainSubstring("-- name: up"))
			Expect(script).To(ContainSubstring("upgrade"))
			Expect(script).To(ContainSubstring("-- name: down"))
			Expect(script).To(ContainSubstring("rollback"))
		})

		Context("when the dir is not valid", func() {
			It("returns an error", func() {
				content := &migration.Content{
					UpCommand:   bytes.NewBufferString("upgrade"),
					DownCommand: bytes.NewBufferString("rollback"),
				}
				generator.FileSystem = parcel.Dir("/")
				Expect(generator.Write(item, content)).To(MatchError("open /20160102150_schema.sql: permission denied"))
			})
		})

		Context("when the up step cannot be created", func() {
			It("returns an error", func() {
				reader := &fake.Buffer{}
				content := &migration.Content{
					UpCommand:   reader,
					DownCommand: bytes.NewBufferString("rollback"),
				}
				reader.ReadReturns(0, fmt.Errorf("Oh no!"))
				Expect(generator.Write(item, content)).To(MatchError("Oh no!"))
			})
		})

		Context("when the up step cannot be created", func() {
			It("returns an error", func() {
				reader := &fake.Buffer{}
				content := &migration.Content{
					UpCommand:   bytes.NewBufferString("upgrade"),
					DownCommand: reader,
				}
				reader.ReadReturns(0, fmt.Errorf("Oh no!"))
				Expect(generator.Write(item, content)).To(MatchError("Oh no!"))
			})
		})
	})
})
