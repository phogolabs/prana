package sqlmigr_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/phogolabs/prana/sqlmigr"
	"github.com/phogolabs/prana/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		Expect(os.MkdirAll(dir, 0700)).To(Succeed())

		generator = &sqlmigr.Generator{
			FileSystem: storage.New(dir),
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
	})
})
