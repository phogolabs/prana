package gom_test

import (
	"bytes"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/svett/gom"
)

var _ = Describe("Embedded", func() {
	Describe("EmbeddedStmt", func() {
		It("prepares the statement correctly", func() {
			stmt := &gom.EmbeddedStmt{
				Query:  "SELECT * FROM users WHERE id = ?",
				Params: []interface{}{1},
			}

			query, params := stmt.Prepare()
			Expect(query).To(Equal("SELECT * FROM users WHERE id = :arg0"))
			Expect(params).To(HaveKeyWithValue("arg0", 1))
		})
	})

	Describe("StmtProvider", func() {
		var provider *gom.StmtProvider

		BeforeEach(func() {
			provider = &gom.StmtProvider{
				Repository: map[string]string{},
			}
		})

		Describe("Load", func() {
			var buffer *bytes.Buffer

			BeforeEach(func() {
				buffer = bytes.NewBufferString("-- name: up")
				fmt.Fprintln(buffer)
				fmt.Fprintln(buffer, "SELECT * FROM users;")
				Expect(provider.Load(buffer)).To(Succeed())
			})

			It("loads the provider successfully", func() {
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

					Expect(provider.Load(buffer)).To(MatchError("Statement 'up' already exists"))
				})
			})
		})

		Describe("Statement", func() {
			BeforeEach(func() {
				buffer := bytes.NewBufferString("-- name: up")
				fmt.Fprintln(buffer)
				fmt.Fprintln(buffer, "SELECT * FROM users")

				Expect(provider.Load(buffer)).To(Succeed())
			})

			It("returns a statement", func() {
				stmt := provider.Statement("up")
				Expect(stmt).NotTo(BeNil())
				Expect(stmt.Params).To(BeEmpty())
				Expect(stmt.Query).To(Equal("SELECT * FROM users"))
			})

			It("returns a statement with params", func() {
				stmt := provider.Statement("up", 1)
				Expect(stmt).NotTo(BeNil())
				Expect(stmt.Params).To(ContainElement(1))
				Expect(stmt.Query).To(Equal("SELECT * FROM users"))
			})

			Context("when not statements are found", func() {
				It("returns a nil statement", func() {
					Expect(provider.Statement("down")).To(BeNil())
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

		It("returns a statement", func() {
			stmt := gom.Statement(script)
			Expect(stmt).NotTo(BeNil())
			Expect(stmt.Params).To(BeEmpty())
			Expect(stmt.Query).To(Equal("SELECT * FROM users"))
		})

		It("returns a statement with params", func() {
			stmt := gom.Statement(script, 1)
			Expect(stmt).NotTo(BeNil())
			Expect(stmt.Params).To(ContainElement(1))
			Expect(stmt.Query).To(Equal("SELECT * FROM users"))
		})

		Context("when the statement does not exits", func() {
			It("does not return a statement", func() {
				Expect(gom.Statement("down")).To(BeNil())
			})
		})
	})
})
