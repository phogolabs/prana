package sqlexec_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/prana/fake"
	"github.com/phogolabs/prana/sqlexec"
	"github.com/phogolabs/parcello"
)

var _ = Describe("Runner", func() {
	var (
		runner *sqlexec.Runner
		dir    string
	)

	BeforeEach(func() {
		var err error

		dir, err = ioutil.TempDir("", "prana_runner")
		Expect(err).To(BeNil())

		db := filepath.Join(dir, "prana.db")
		gateway, err := sqlx.Open("sqlite3", db)
		Expect(err).To(BeNil())

		runner = &sqlexec.Runner{
			FileSystem: parcello.Dir(dir),
			DB:         gateway,
		}
	})

	JustBeforeEach(func() {
		command := &bytes.Buffer{}
		fmt.Fprintln(command, "-- name: system-tables")
		fmt.Fprintln(command, "SELECT * FROM sqlite_master")

		path := filepath.Join(dir, "commands.sql")
		Expect(ioutil.WriteFile(path, command.Bytes(), 0700)).To(Succeed())
	})

	AfterEach(func() {
		runner.DB.Close()
	})

	It("runs the command successfully", func() {
		rows, err := runner.Run("system-tables")
		Expect(err).To(Succeed())

		columns, err := rows.Columns()
		Expect(err).To(Succeed())
		Expect(columns).To(ContainElement("type"))
		Expect(columns).To(ContainElement("name"))
		Expect(columns).To(ContainElement("tbl_name"))
		Expect(columns).To(ContainElement("rootpage"))
		Expect(columns).To(ContainElement("sql"))
	})

	Context("when the file system fails", func() {
		BeforeEach(func() {
			fileSystem := &fake.FileSystem{}
			fileSystem.WalkReturns(fmt.Errorf("Oh no!"))
			runner.FileSystem = fileSystem
		})

		It("returns the error", func() {
			_, err := runner.Run("system-tables")
			Expect(err).To(MatchError("Oh no!"))
		})
	})

	Context("when the command requires parameters", func() {
		JustBeforeEach(func() {
			command := &bytes.Buffer{}
			fmt.Fprintln(command, "-- name: system-tables")
			fmt.Fprintln(command, "SELECT ? AS Param FROM sqlite_master")

			path := filepath.Join(dir, "commands.sql")
			Expect(ioutil.WriteFile(path, command.Bytes(), 0700)).To(Succeed())
		})

		It("runs the command successfully", func() {
			rows, err := runner.Run("system-tables", "hello")
			Expect(err).To(Succeed())

			columns, err := rows.Columns()
			Expect(err).To(Succeed())
			Expect(columns).To(ContainElement("Param"))
		})
	})

	Context("when the command does not exist", func() {
		JustBeforeEach(func() {
			path := filepath.Join(dir, "commands.sql")
			Expect(os.Remove(path)).To(Succeed())
		})

		It("returns an error", func() {
			_, err := runner.Run("system-tables")
			Expect(err).To(MatchError("Command 'system-tables' not found"))
		})
	})

	Context("when the database is not available", func() {
		JustBeforeEach(func() {
			Expect(runner.DB.Close()).To(Succeed())
		})

		It("return an error", func() {
			_, err := runner.Run("system-tables")
			Expect(err).To(MatchError("sql: database is closed"))
		})
	})
})
