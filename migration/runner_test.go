package migration_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/gom"
	"github.com/phogolabs/gom/migration"
)

var _ = Describe("Runner", func() {
	var (
		runner *migration.Runner
		item   *migration.Item
	)

	BeforeEach(func() {
		dir, err := ioutil.TempDir("", "gom_runner")
		Expect(err).To(BeNil())

		db := filepath.Join(dir, "gom.db")
		gateway, err := gom.Open("sqlite3", db)
		Expect(err).To(BeNil())

		runner = &migration.Runner{
			Dir:     dir,
			Gateway: gateway,
		}

		item = &migration.Item{
			Id:          "20160102150",
			Description: "schema",
		}
	})

	JustBeforeEach(func() {
		query := &bytes.Buffer{}
		fmt.Fprintln(query, "CREATE TABLE migrations (")
		fmt.Fprintln(query, " id          TEXT      NOT NULL PRIMARY KEY,")
		fmt.Fprintln(query, " description TEXT      NOT NULL,")
		fmt.Fprintln(query, " created_at  TIMESTAMP NOT NULL")
		fmt.Fprintln(query, ");")

		_, err := runner.Gateway.DB().Exec(query.String())
		Expect(err).To(BeNil())

		migration := &bytes.Buffer{}
		fmt.Fprintln(migration, "-- name: up")
		fmt.Fprintln(migration, "CREATE TABLE test(id TEXT)")
		fmt.Fprintln(migration, "-- name: down")
		fmt.Fprintln(migration, "DROP TABLE IF EXISTS test")

		path := filepath.Join(runner.Dir, item.Filename())
		Expect(ioutil.WriteFile(path, migration.Bytes(), 0700)).To(Succeed())
	})

	AfterEach(func() {
		runner.Gateway.Close()
	})

	Describe("Run", func() {
		It("runs the migration successfully", func() {
			Expect(runner.Run(item)).To(Succeed())
			_, err := runner.Gateway.Exec(gom.Select("id").From(gom.Table("test")))
			Expect(err).NotTo(HaveOccurred())

			items := []migration.Item{}
			query := gom.Select().From("migrations")

			Expect(runner.Gateway.Select(&items, query)).To(Succeed())
			Expect(items).To(HaveLen(1))

			Expect(items[0].Id).To(Equal(item.Id))
			Expect(items[0].Description).To(Equal(item.Description))
		})

		Context("when the migration does not exist", func() {
			JustBeforeEach(func() {
				path := filepath.Join(runner.Dir, item.Filename())
				Expect(os.Remove(path)).To(Succeed())
			})

			It("returns an error", func() {
				path := filepath.Join(runner.Dir, item.Filename())
				msg := fmt.Sprintf("open %s: no such file or directory", path)
				Expect(runner.Run(item)).To(MatchError(msg))
			})
		})

		Context("when the database is not available", func() {
			JustBeforeEach(func() {
				Expect(runner.Gateway.Close()).To(Succeed())
			})

			It("return an error", func() {
				Expect(runner.Run(item)).To(MatchError("sql: database is closed"))
			})
		})

		Context("when the migration step does not exist", func() {
			JustBeforeEach(func() {
				migration := &bytes.Buffer{}
				fmt.Fprintln(migration, "-- name: down")
				fmt.Fprintln(migration, "DROP TABLE IF EXISTS test")

				path := filepath.Join(runner.Dir, item.Filename())
				Expect(ioutil.WriteFile(path, migration.Bytes(), 0700)).To(Succeed())
			})

			It("return an error", func() {
				Expect(runner.Run(item)).To(MatchError("Command 'up' not found"))
			})
		})

		Context("when the dir is not valid", func() {
			It("returns an error", func() {
				runner.Dir = ""
				Expect(runner.Run(item)).To(MatchError("open 20160102150_schema.sql: no such file or directory"))
			})
		})

		Context("when the item is run more than once", func() {
			It("return an error", func() {
				Expect(runner.Run(item)).To(Succeed())
				_, err := runner.Gateway.DB().Exec("DROP TABLE test")
				Expect(err).NotTo(HaveOccurred())
				Expect(runner.Run(item)).To(MatchError("UNIQUE constraint failed: migrations.id"))
			})
		})
	})

	Describe("Revert", func() {
		It("reverts the migration successfully", func() {
			Expect(runner.Revert(item)).To(Succeed())
			_, err := runner.Gateway.Exec(gom.Select("id").From(gom.Table("test")))
			Expect(err).To(MatchError("no such table: test"))

			items := []migration.Item{}
			query := gom.Select().From("migrations")

			Expect(runner.Gateway.Select(&items, query)).To(Succeed())
			Expect(items).To(HaveLen(0))
		})

		Context("when the migration does not exist", func() {
			JustBeforeEach(func() {
				path := filepath.Join(runner.Dir, item.Filename())
				Expect(os.Remove(path)).To(Succeed())
			})

			It("returns an error", func() {
				path := filepath.Join(runner.Dir, item.Filename())
				msg := fmt.Sprintf("open %s: no such file or directory", path)
				Expect(runner.Revert(item)).To(MatchError(msg))
			})
		})

		Context("when the database is not available", func() {
			JustBeforeEach(func() {
				Expect(runner.Gateway.Close()).To(Succeed())
			})

			It("return an error", func() {
				Expect(runner.Revert(item)).To(MatchError("sql: database is closed"))
			})
		})

		Context("when the migration step does not exist", func() {
			JustBeforeEach(func() {
				migration := &bytes.Buffer{}
				fmt.Fprintln(migration, "-- name: up")
				fmt.Fprintln(migration, "CREATE TABLE test(id TEXT)")

				path := filepath.Join(runner.Dir, item.Filename())
				Expect(ioutil.WriteFile(path, migration.Bytes(), 0700)).To(Succeed())
			})

			It("return an error", func() {
				Expect(runner.Revert(item)).To(MatchError("Command 'down' not found"))
			})
		})

		Context("when the dir is not valid", func() {
			It("returns an error", func() {
				runner.Dir = ""
				Expect(runner.Revert(item)).To(MatchError("open 20160102150_schema.sql: no such file or directory"))
			})
		})

		Context("when the migration table is locked", func() {
			It("return an error", func() {
				tx, err := runner.Gateway.DB().Begin()
				Expect(err).NotTo(HaveOccurred())

				_, err = tx.Exec("SELECT * FROM migrations")
				Expect(err).NotTo(HaveOccurred())

				Expect(runner.Revert(item)).To(MatchError("database is locked"))
				Expect(tx.Commit()).To(Succeed())
			})
		})
	})
})
