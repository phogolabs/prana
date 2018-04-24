package sqlexec_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/prana/fake"
	"github.com/phogolabs/prana/sqlexec"
	"github.com/phogolabs/parcello"
)

var _ = Describe("Generator", func() {
	var (
		generator *sqlexec.Generator
		dir       string
	)

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "prana_generator")
		Expect(err).To(BeNil())

		generator = &sqlexec.Generator{
			FileSystem: parcello.Dir(dir),
		}
	})

	Describe("Create", func() {
		It("creates a command file successfully", func() {
			name, path, err := generator.Create("commands", "update")
			Expect(err).To(BeNil())

			path = filepath.Join(dir, path)
			Expect(path).To(BeARegularFile())
			Expect(name).To(Equal("update"))

			data, err := ioutil.ReadFile(path)
			Expect(err).To(BeNil())

			sqlexec := string(data)
			Expect(sqlexec).To(ContainSubstring("-- name: update"))
		})

		Context("when the file is not provided", func() {
			It("creates a command file successfully", func() {
				name, path, err := generator.Create("", "update")
				Expect(err).To(BeNil())

				path = filepath.Join(dir, path)
				Expect(path).To(BeARegularFile())
				Expect(name).To(Equal("update"))

				filename := filepath.Base(path)
				ext := filepath.Ext(path)
				filename = strings.Replace(filename, ext, "", -1)

				_, err = time.Parse("20060102150405", filename)
				Expect(err).To(Succeed())

				data, err := ioutil.ReadFile(path)
				Expect(err).To(BeNil())

				sqlexec := string(data)
				Expect(sqlexec).To(ContainSubstring("-- name: update"))
			})
		})

		Context("when the file already exists", func() {
			It("adds the command to the file successfully", func() {
				name, path, err := generator.Create("commands", "update")
				Expect(err).To(BeNil())
				Expect(name).To(Equal("update"))

				path = filepath.Join(dir, path)
				Expect(path).To(BeARegularFile())

				name, path, err = generator.Create("commands", "delete")
				Expect(err).To(BeNil())
				Expect(name).To(Equal("delete"))

				path = filepath.Join(dir, path)
				Expect(path).To(BeARegularFile())

				data, err := ioutil.ReadFile(path)
				Expect(err).To(BeNil())

				sqlexec := string(data)
				Expect(sqlexec).To(ContainSubstring("-- name: update"))
				Expect(sqlexec).To(ContainSubstring("-- name: delete"))
			})
		})

		Context("when opening a file fails", func() {
			BeforeEach(func() {
				fileSystem := &fake.FileSystem{}
				fileSystem.OpenFileReturns(nil, fmt.Errorf("Oh no!"))
				generator.FileSystem = fileSystem
			})

			It("returns an error", func() {
				_, _, err := generator.Create("commands", "update")
				Expect(err).To(MatchError("Oh no!"))
			})
		})

		Context("when the command already exists", func() {
			It("returns an error", func() {
				buffer := &bytes.Buffer{}
				fmt.Fprintln(buffer, "-- name: update")
				fmt.Fprintln(buffer, "SELECT * FROM migrations;")

				path := filepath.Join(dir, "commands.sql")
				Expect(ioutil.WriteFile(path, buffer.Bytes(), 0700)).To(Succeed())

				_, _, err := generator.Create("commands", "update")
				Expect(err).To(MatchError("Command 'update' already exists"))
			})
		})

		Context("when the dir is not valid", func() {
			It("returns an error", func() {
				generator.FileSystem = parcello.Dir("/hello")
				_, _, err := generator.Create("commands", "update")
				Expect(err).To(MatchError("Directory does not exist"))
			})
		})
	})
})
