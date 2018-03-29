package script_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/svett/gom/script"
)

var _ = Describe("Provider", func() {
	var provider *script.Provider

	BeforeEach(func() {
		provider = &script.Provider{}
	})

	Describe("Load", func() {
		var buffer *bytes.Buffer

		BeforeEach(func() {
			buffer = bytes.NewBufferString("-- name: up")
			fmt.Fprintln(buffer)
			fmt.Fprintln(buffer, "SELECT * FROM users;")
		})

		It("loads the provider successfully", func() {
			Expect(provider.Load(buffer)).To(Succeed())

			cmd, err := provider.Command("up")
			Expect(err).To(BeNil())

			Expect(cmd.Query).To(Equal("SELECT * FROM users;"))
		})

		Context("when the statement are duplicated", func() {
			It("returns an error", func() {
				Expect(provider.Load(buffer)).To(Succeed())

				buffer = bytes.NewBufferString("-- name: up")
				fmt.Fprintln(buffer)
				fmt.Fprintln(buffer, "SELECT * FROM categories;")

				Expect(provider.Load(buffer)).To(MatchError("Command 'up' already exists"))
			})
		})
	})

	Describe("LoadDir", func() {
		var dir string

		BeforeEach(func() {
			var err error

			dir, err = ioutil.TempDir("", "gom_generator")
			Expect(err).To(BeNil())

			path := filepath.Join(dir, "commands.sql")

			buffer := bytes.NewBufferString("-- name: up")
			fmt.Fprintln(buffer)
			fmt.Fprintln(buffer, "SELECT * FROM users;")

			Expect(ioutil.WriteFile(path, buffer.Bytes(), 0700)).To(Succeed())
		})

		It("loads the provider successfully", func() {
			Expect(provider.LoadDir(dir)).To(Succeed())

			cmd, err := provider.Command("up")
			Expect(err).To(BeNil())

			Expect(cmd.Query).To(Equal("SELECT * FROM users;"))
		})

		Context("when the statement are duplicated", func() {
			It("returns an error", func() {
				path := filepath.Join(dir, "another.sql")

				buffer := bytes.NewBufferString("-- name: up")
				fmt.Fprintln(buffer)
				fmt.Fprintln(buffer, "SELECT * FROM users;")

				Expect(ioutil.WriteFile(path, buffer.Bytes(), 0700)).To(Succeed())
				Expect(provider.LoadDir(dir)).To(MatchError("Command 'up' already exists"))
			})
		})
	})

	Describe("Command", func() {
		BeforeEach(func() {
			buffer := bytes.NewBufferString("-- name: up")
			fmt.Fprintln(buffer)
			fmt.Fprintln(buffer, "SELECT * FROM users")

			Expect(provider.Load(buffer)).To(Succeed())
		})

		It("returns a command", func() {
			stmt, err := provider.Command("up")
			Expect(err).To(BeNil())
			Expect(stmt).NotTo(BeNil())
			Expect(stmt.Params).To(BeEmpty())
			Expect(stmt.Query).To(Equal("SELECT * FROM users"))
		})

		It("returns a command with params", func() {
			stmt, err := provider.Command("up", 1)
			Expect(err).To(BeNil())
			Expect(stmt).NotTo(BeNil())
			Expect(stmt.Params).To(ContainElement(1))
			Expect(stmt.Query).To(Equal("SELECT * FROM users"))
		})

		Context("when not statements are found", func() {
			It("returns a error", func() {
				cmd, err := provider.Command("down")
				Expect(err).To(MatchError("Command 'down' not found"))
				Expect(cmd).To(BeNil())
			})
		})
	})
})
