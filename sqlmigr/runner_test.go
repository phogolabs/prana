package sqlmigr_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/parcello"
	"github.com/phogolabs/prana/sqlmigr"
)

var _ = Describe("Runner", func() {
	var (
		runner *sqlmigr.Runner
		item   *sqlmigr.Migration
		dir    string
	)

	BeforeEach(func() {
		var err error

		dir, err = ioutil.TempDir("", "prana_runner")
		Expect(err).To(BeNil())

		conn := filepath.Join(dir, "prana.db")
		db, err := sqlx.Open("sqlite3", conn)
		Expect(err).To(BeNil())

		runner = &sqlmigr.Runner{
			FileSystem: parcello.Dir(dir),
			DB:         db,
		}

		item = &sqlmigr.Migration{
			ID:          "20160102150",
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

		_, err := runner.DB.Exec(query.String())
		Expect(err).To(BeNil())

		sqlmigr := &bytes.Buffer{}
		fmt.Fprintln(sqlmigr, "-- name: up")
		fmt.Fprintln(sqlmigr, "CREATE TABLE test(id TEXT);")
		fmt.Fprintln(sqlmigr, "CREATE TABLE test2(id TEXT);")
		fmt.Fprintln(sqlmigr, "-- name: down")
		fmt.Fprintln(sqlmigr, "DROP TABLE IF EXISTS test;")
		fmt.Fprintln(sqlmigr, "DROP TABLE IF EXISTS test2;")

		path := filepath.Join(dir, item.Filename())
		Expect(ioutil.WriteFile(path, sqlmigr.Bytes(), 0700)).To(Succeed())
	})

	AfterEach(func() {
		runner.DB.Close()
	})

	Describe("Run", func() {
		It("runs the sqlmigr successfully", func() {
			Expect(runner.Run(item)).To(Succeed())
			_, err := runner.DB.Exec("SELECT id FROM test")
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the sqlmigr does not exist", func() {
			JustBeforeEach(func() {
				path := filepath.Join(dir, item.Filename())
				Expect(os.Remove(path)).To(Succeed())
			})

			It("returns an error", func() {
				path := filepath.Join(dir, item.Filename())
				msg := fmt.Sprintf("open %s: no such file or directory", path)
				Expect(runner.Run(item)).To(MatchError(msg))
			})
		})

		Context("when the database is not available", func() {
			JustBeforeEach(func() {
				Expect(runner.DB.Close()).To(Succeed())
			})

			It("return an error", func() {
				Expect(runner.Run(item)).To(MatchError("sql: database is closed"))
			})
		})

		Context("when the sqlmigr step does not exist", func() {
			JustBeforeEach(func() {
				sqlmigr := &bytes.Buffer{}
				fmt.Fprintln(sqlmigr, "-- name: down")
				fmt.Fprintln(sqlmigr, "DROP TABLE IF EXISTS test")

				path := filepath.Join(dir, item.Filename())
				Expect(ioutil.WriteFile(path, sqlmigr.Bytes(), 0700)).To(Succeed())
			})

			It("return an error", func() {
				Expect(runner.Run(item)).To(MatchError("Command 'up' not found for migration '20160102150_schema.sql'"))
			})
		})

		Context("when the dir is not valid", func() {
			It("returns an error", func() {
				runner.FileSystem = parcello.Dir("/")
				Expect(runner.Run(item).Error()).To(Equal("open /20160102150_schema.sql: no such file or directory"))
			})
		})
	})

	Describe("Revert", func() {
		It("reverts the migration successfully", func() {
			Expect(runner.Revert(item)).To(Succeed())
			_, err := runner.DB.Exec("SELECT id FROM test")
			Expect(err).To(MatchError("no such table: test"))
		})

		Context("when the migration does not exist", func() {
			JustBeforeEach(func() {
				path := filepath.Join(dir, item.Filename())
				Expect(os.Remove(path)).To(Succeed())
			})

			It("returns an error", func() {
				path := filepath.Join(dir, item.Filename())
				msg := fmt.Sprintf("open %s: no such file or directory", path)
				Expect(runner.Revert(item)).To(MatchError(msg))
			})
		})

		Context("when the database is not available", func() {
			JustBeforeEach(func() {
				Expect(runner.DB.Close()).To(Succeed())
			})

			It("return an error", func() {
				Expect(runner.Revert(item)).To(MatchError("sql: database is closed"))
			})
		})

		Context("when the sqlmigr step does not exist", func() {
			JustBeforeEach(func() {
				sqlmigr := &bytes.Buffer{}
				fmt.Fprintln(sqlmigr, "-- name: up")
				fmt.Fprintln(sqlmigr, "CREATE TABLE test(id TEXT)")

				path := filepath.Join(dir, item.Filename())
				Expect(ioutil.WriteFile(path, sqlmigr.Bytes(), 0700)).To(Succeed())
			})

			It("return an error", func() {
				Expect(runner.Revert(item)).To(MatchError("Command 'down' not found for migration '20160102150_schema.sql'"))
			})
		})

		Context("when the dir is not valid", func() {
			It("returns an error", func() {
				runner.FileSystem = parcello.Dir("/")
				Expect(runner.Revert(item)).To(MatchError("open /20160102150_schema.sql: no such file or directory"))
			})
		})
	})
})
