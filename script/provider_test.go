package script_test

import (
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/phogolabs/oak/fake"
	"github.com/phogolabs/oak/script"
)

var _ = Describe("Provider", func() {
	var provider *script.Provider

	BeforeEach(func() {
		provider = &script.Provider{}
	})

	Describe("ReadFrom", func() {
		var buffer *bytes.Buffer

		BeforeEach(func() {
			buffer = bytes.NewBufferString("-- name: up")
			fmt.Fprintln(buffer)
			fmt.Fprintln(buffer, "SELECT * FROM users;")
		})

		It("loads the provider successfully", func() {
			n, err := provider.ReadFrom(buffer)
			Expect(n).To(Equal(int64(1)))
			Expect(err).To(Succeed())

			cmd, err := provider.Command("up")
			Expect(err).To(BeNil())

			query, _ := cmd.Prepare()
			Expect(query).To(Equal("SELECT * FROM users;"))
		})

		Context("when the statement are duplicated", func() {
			It("returns an error", func() {
				n, err := provider.ReadFrom(buffer)
				Expect(n).To(Equal(int64(1)))
				Expect(err).To(Succeed())

				buffer = bytes.NewBufferString("-- name: up")
				fmt.Fprintln(buffer)
				fmt.Fprintln(buffer, "SELECT * FROM categories;")

				n, err = provider.ReadFrom(buffer)
				Expect(n).To(BeZero())
				Expect(err).To(MatchError("Command 'up' already exists"))
			})
		})
	})

	Describe("ReadDir", func() {
		var fileSystem *fake.FileSystem

		BeforeEach(func() {
			fileSystem = &fake.FileSystem{}
		})

		It("loads the provider successfully", func() {
			Expect(provider.ReadDir(fileSystem)).To(Succeed())
			Expect(fileSystem.WalkCallCount()).To(Equal(1))

			dir, _ := fileSystem.WalkArgsForCall(0)
			Expect(dir).To(Equal("/"))
		})

		Context("when the file system fails ", func() {
			It("returns an error", func() {
				fileSystem.WalkReturns(fmt.Errorf("Oh no!"))
				Expect(provider.ReadDir(fileSystem)).To(MatchError("Oh no!"))
			})
		})
	})

	Describe("Command", func() {
		BeforeEach(func() {
			buffer := bytes.NewBufferString("-- name: up")
			fmt.Fprintln(buffer)
			fmt.Fprintln(buffer, "SELECT * FROM users")

			n, err := provider.ReadFrom(buffer)
			Expect(n).To(Equal(int64(1)))
			Expect(err).To(Succeed())
		})

		It("returns a command", func() {
			stmt, err := provider.Command("up")
			Expect(err).To(BeNil())
			Expect(stmt).NotTo(BeNil())

			query, params := stmt.Prepare()
			Expect(params).To(BeEmpty())
			Expect(query).To(Equal("SELECT * FROM users"))
		})

		Context("when the command has arguments", func() {
			BeforeEach(func() {
				buffer := bytes.NewBufferString("-- name: show-users")
				fmt.Fprintln(buffer)
				fmt.Fprintln(buffer, "SELECT * FROM users WHERE id = ?")

				n, err := provider.ReadFrom(buffer)
				Expect(n).To(Equal(int64(1)))
				Expect(err).To(Succeed())
			})

			It("returns a command with params", func() {
				stmt, err := provider.Command("show-users", 1)
				Expect(err).To(BeNil())
				Expect(stmt).NotTo(BeNil())

				query, params := stmt.Prepare()
				Expect(query).To(Equal("SELECT * FROM users WHERE id = :arg0"))
				Expect(params).To(HaveKeyWithValue("arg0", 1))
			})
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
