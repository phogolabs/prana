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
	"github.com/phogolabs/prana/fake"
	"github.com/phogolabs/prana/sqlmigr"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
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
			Drivers:     []string{"sql"},
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
		fmt.Fprintln(sqlmigr, "CREATE TABLE IF NOT EXISTS test(id TEXT);")
		fmt.Fprintln(sqlmigr, "CREATE TABLE IF NOT EXISTS test2(id TEXT);")
		fmt.Fprintln(sqlmigr, "-- name: down")
		fmt.Fprintln(sqlmigr, "DROP TABLE IF EXISTS test;")
		fmt.Fprintln(sqlmigr, "DROP TABLE IF EXISTS test2;")

		for _, filename := range item.Filenames() {
			path := filepath.Join(dir, filename)
			Expect(ioutil.WriteFile(path, sqlmigr.Bytes(), 0700)).To(Succeed())
		}
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
				for _, filename := range item.Filenames() {
					path := filepath.Join(dir, filename)
					Expect(os.Remove(path)).To(Succeed())
				}
			})

			It("returns an error", func() {
				path := filepath.Join(dir, item.Filenames()[0])
				msg := fmt.Sprintf("open %s: no such file or directory", path)
				Expect(runner.Run(item)).To(MatchError(msg))
			})
		})

		Context("when the database is not available", func() {
			JustBeforeEach(func() {
				Expect(runner.DB.Close()).To(Succeed())
			})

			It("return an error", func() {
				err := runner.Run(item)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("sql: database is closed"))
			})
		})

		Context("when the command execution fails", func() {
			var mock sqlmock.Sqlmock

			JustBeforeEach(func() {
				db, m, err := sqlmock.New()
				Expect(err).NotTo(HaveOccurred())
				runner.DB = sqlx.NewDb(db, "dummy")

				mock = m
				mock.ExpectBegin()
				mock.ExpectExec("CREATE TABLE test(id TEXT)").
					WillReturnError(fmt.Errorf("oh no!"))
				mock.ExpectRollback()
			})

			It("returns the error", func() {
				Expect(runner.Run(item)).To(HaveOccurred())
			})
		})

		Context("when the sqlmigr step does not exist", func() {
			JustBeforeEach(func() {
				sqlmigr := &bytes.Buffer{}
				fmt.Fprintln(sqlmigr, "-- name: down")
				fmt.Fprintln(sqlmigr, "DROP TABLE IF EXISTS test")

				path := filepath.Join(dir, item.Filenames()[0])
				Expect(ioutil.WriteFile(path, sqlmigr.Bytes(), 0700)).To(Succeed())
			})

			It("return an error", func() {
				Expect(runner.Run(item)).To(HaveOccurred())
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

		Context("when the migration has multiple files", func() {
			BeforeEach(func() {
				item.Drivers = []string{"sqlite3", "sql"}
			})

			It("reverts the migration successfully", func() {
				Expect(runner.Revert(item)).To(Succeed())
				_, err := runner.DB.Exec("SELECT id FROM test")
				Expect(err).To(MatchError("no such table: test"))
			})
		})

		Context("when the migration does not exist", func() {
			JustBeforeEach(func() {
				path := filepath.Join(dir, item.Filenames()[0])
				Expect(os.Remove(path)).To(Succeed())
			})

			It("returns an error", func() {
				path := filepath.Join(dir, item.Filenames()[0])
				msg := fmt.Sprintf("open %s: no such file or directory", path)
				Expect(runner.Revert(item)).To(MatchError(msg))
			})
		})

		Context("when the database is not available", func() {
			JustBeforeEach(func() {
				Expect(runner.DB.Close()).To(Succeed())
			})

			It("return an error", func() {
				err := runner.Revert(item)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("sql: database is closed"))
			})
		})

		Context("when the sqlmigr step does not exist", func() {
			JustBeforeEach(func() {
				sqlmigr := &bytes.Buffer{}
				fmt.Fprintln(sqlmigr, "-- name: up")
				fmt.Fprintln(sqlmigr, "CREATE TABLE test(id TEXT)")

				path := filepath.Join(dir, item.Filenames()[0])
				Expect(ioutil.WriteFile(path, sqlmigr.Bytes(), 0700)).To(Succeed())
			})

			It("return an error", func() {
				Expect(runner.Revert(item)).To(MatchError("routine 'down' not found for migration '20160102150_schema'"))
			})
		})

		Context("when the dir is not valid", func() {
			JustBeforeEach(func() {
				fs := &fake.FileSystem{}
				fs.OpenFileReturns(nil, fmt.Errorf("oh no!"))
				runner.FileSystem = fs
			})

			It("returns an error", func() {
				Expect(runner.Revert(item)).To(MatchError("oh no!"))
			})
		})
	})
})
