package migration_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/gom/fake"
	"github.com/phogolabs/gom/migration"
)

var _ = Describe("Generator", func() {
	var (
		generator *migration.Generator
		item      *migration.Item
	)

	BeforeEach(func() {
		dir, err := ioutil.TempDir("", "gom_generator")
		Expect(err).To(BeNil())

		dir = filepath.Join(dir, "migration")

		generator = &migration.Generator{
			Dir: dir,
		}

		item = &migration.Item{
			Id:          "20160102150",
			Description: "schema",
		}
	})

	Describe("Create", func() {
		It("creates a migration successfully", func() {
			path, err := generator.Create(item)
			Expect(err).To(BeNil())
			Expect(path).To(BeARegularFile())
			Expect(path).To(Equal(filepath.Join(generator.Dir, item.Filename())))
			Expect(generator.Dir).To(BeADirectory())

			data, err := ioutil.ReadFile(path)
			Expect(err).To(BeNil())

			script := string(data)
			Expect(script).To(ContainSubstring("-- name: up"))
			Expect(script).To(ContainSubstring("-- name: down"))
		})

		Context("when the migration exists", func() {
			It("returns an error", func() {
				var err error

				path := filepath.Join(generator.Dir, item.Filename())
				msg := fmt.Sprintf("Migration '%s' already exists", path)

				_, err = generator.Create(item)
				Expect(err).To(BeNil())

				_, err = generator.Create(item)
				Expect(err).To(MatchError(msg))
			})
		})

		Context("when the dir is not valid", func() {
			It("returns an error", func() {
				generator.Dir = ""
				_, err := generator.Create(item)
				Expect(err).To(MatchError("mkdir : no such file or directory"))
			})
		})

		Context("when the dir is the root dir", func() {
			It("returns an error", func() {
				generator.Dir = "/"
				_, err := generator.Create(item)
				Expect(err).To(MatchError("open /20160102150_schema.sql: permission denied"))
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
			Expect(generator.Dir).To(BeADirectory())

			path := filepath.Join(generator.Dir, item.Filename())
			data, err := ioutil.ReadFile(path)
			Expect(err).To(BeNil())

			script := string(data)
			Expect(script).To(ContainSubstring("-- name: up"))
			Expect(script).To(ContainSubstring("upgrade"))
			Expect(script).To(ContainSubstring("-- name: down"))
			Expect(script).To(ContainSubstring("rollback"))
		})

		Context("when the migration exists", func() {
			It("returns an error", func() {
				content := &migration.Content{
					UpCommand:   bytes.NewBufferString("upgrade"),
					DownCommand: bytes.NewBufferString("rollback"),
				}

				Expect(generator.Write(item, content)).To(Succeed())

				path := filepath.Join(generator.Dir, item.Filename())
				msg := fmt.Sprintf("Migration '%s' already exists", path)
				Expect(generator.Write(item, content)).To(MatchError(msg))
			})
		})

		Context("when the dir is not valid", func() {
			It("returns an error", func() {
				content := &migration.Content{
					UpCommand:   bytes.NewBufferString("upgrade"),
					DownCommand: bytes.NewBufferString("rollback"),
				}
				generator.Dir = ""
				Expect(generator.Write(item, content)).To(MatchError("mkdir : no such file or directory"))
			})
		})

		Context("when the up step cannot be created", func() {
			It("returns an error", func() {
				reader := &fake.Reader{}
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
				reader := &fake.Reader{}
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
