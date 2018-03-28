package gom_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/svett/gom"
)

var _ = Describe("Command", func() {
	Describe("Cmd", func() {
		It("prepares the command correctly", func() {
			stmt := &gom.Cmd{
				Query:  "SELECT * FROM users WHERE id = ?",
				Params: []gom.Param{1},
			}

			query, params := stmt.Prepare()
			Expect(query).To(Equal("SELECT * FROM users WHERE id = :arg0"))
			Expect(params).To(HaveKeyWithValue("arg0", 1))
		})
	})

	Describe("CmdProvider", func() {
		var provider *gom.CmdProvider

		BeforeEach(func() {
			provider = &gom.CmdProvider{
				Repository: map[string]string{},
			}
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

				query, ok := provider.Repository["up"]
				Expect(ok).To(BeTrue())

				Expect(query).To(Equal("SELECT * FROM users;"))
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

				query, ok := provider.Repository["up"]
				Expect(ok).To(BeTrue())

				Expect(query).To(Equal("SELECT * FROM users;"))
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

	Describe("Statement", func() {
		var script string

		BeforeEach(func() {
			script = fmt.Sprintf("%v", time.Now().UnixNano())
			buffer := bytes.NewBufferString(fmt.Sprintf("-- name: %v", script))
			fmt.Fprintln(buffer)
			fmt.Fprintln(buffer, "SELECT * FROM users")
			Expect(gom.Load(buffer)).To(Succeed())
		})

		It("returns a command", func() {
			stmt := gom.Command(script)
			Expect(stmt).NotTo(BeNil())
			Expect(stmt.Params).To(BeEmpty())
			Expect(stmt.Query).To(Equal("SELECT * FROM users"))
		})

		It("returns a command with params", func() {
			stmt := gom.Command(script, 1)
			Expect(stmt).NotTo(BeNil())
			Expect(stmt.Params).To(ContainElement(1))
			Expect(stmt.Query).To(Equal("SELECT * FROM users"))
		})

		Context("when the statement does not exits", func() {
			It("does not return a statement", func() {
				Expect(func() { gom.Command("down") }).To(Panic())
			})
		})
	})
})
