package sqlexec_test

import (
	"bytes"
	"fmt"
	"testing/fstest"

	"github.com/phogolabs/prana/sqlexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Provider", func() {
	var provider *sqlexec.Provider

	BeforeEach(func() {
		provider = &sqlexec.Provider{}
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

			query, err := provider.Query("up")
			Expect(err).NotTo(HaveOccurred())
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
				Expect(err).To(MatchError("query 'up' already exists"))
			})
		})
	})

	Describe("ReadDir", func() {
		var storage fstest.MapFS

		BeforeEach(func() {
			storage = fstest.MapFS{}
		})

		It("loads the provider successfully", func() {
			buffer := &bytes.Buffer{}
			fmt.Fprintln(buffer, "-- name: get-categories")
			fmt.Fprintln(buffer, "SELECT * FROM categories;")

			storage["routine.sql"] = &fstest.MapFile{
				Data: buffer.Bytes(),
			}

			Expect(provider.ReadDir(storage)).To(Succeed())

			query, err := provider.Query("get-categories")
			Expect(err).NotTo(HaveOccurred())
			Expect(query).To(Equal("SELECT * FROM categories;"))
		})

		It("skips none sql files", func() {
			buffer := &bytes.Buffer{}
			fmt.Fprintln(buffer, "-- name: get-categories")
			fmt.Fprintln(buffer, "SELECT * FROM categories;")

			storage["routine.txt"] = &fstest.MapFile{
				Data: buffer.Bytes(),
			}

			Expect(provider.ReadDir(storage)).To(Succeed())

			query, err := provider.Query("get-categories")
			Expect(err).To(MatchError("query 'get-categories' not found"))
			Expect(query).To(BeEmpty())
		})
	})
})
