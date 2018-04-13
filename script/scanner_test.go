package script_test

import (
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/oak/script"
)

var _ = Describe("Scanner", func() {
	var scanner *script.Scanner

	BeforeEach(func() {
		scanner = &script.Scanner{}
	})

	It("returns the tagged statements successfully", func() {
		buffer := &bytes.Buffer{}
		fmt.Fprintln(buffer, "-- name: save-user")
		fmt.Fprintln(buffer, "SELECT * FROM users;")

		queries := scanner.Scan(buffer)

		Expect(queries).To(HaveLen(1))
		Expect(queries).To(HaveKeyWithValue("save-user", "SELECT * FROM users;"))
	})

	Context("when the tag is followed by another comment", func() {
		It("returns the tagged statements successfully", func() {
			buffer := &bytes.Buffer{}
			fmt.Fprintln(buffer, "-- name: save-user")
			fmt.Fprintln(buffer, "-- information")
			fmt.Fprintln(buffer, "SELECT * FROM users;")

			queries := scanner.Scan(buffer)

			Expect(queries).To(HaveLen(1))
			Expect(queries).To(HaveKeyWithValue("save-user", "-- information\nSELECT * FROM users;"))
		})
	})

	Context("when there is a tag does not have body", func() {
		It("returns an empty repository", func() {
			buffer := &bytes.Buffer{}
			fmt.Fprintln(buffer, "-- name: empty-query")
			fmt.Fprintln(buffer, "-- name: save-user")
			fmt.Fprintln(buffer, "SELECT * FROM users;")

			queries := scanner.Scan(buffer)

			Expect(queries).To(HaveLen(1))
			Expect(queries).To(HaveKeyWithValue("save-user", "SELECT * FROM users;"))
		})
	})
})
