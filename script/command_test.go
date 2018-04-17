package script_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/oak/script"
)

var _ = Describe("Command", func() {
	It("prepares the command correctly", func() {
		stmt := script.SQL("SELECT * FROM users WHERE id = ?", 1)
		query, params := stmt.Prepare()
		Expect(query).To(Equal("SELECT * FROM users WHERE id = :arg0"))
		Expect(params).To(HaveKeyWithValue("arg0", 1))
	})

	Context("when parameters are more than the required number", func() {
		It("returns the query", func() {
			stmt := script.SQL("SELECT * FROM users WHERE id = ? and name = ?", 2)
			query, params := stmt.Prepare()
			Expect(query).To(BeEmpty())
			Expect(params).To(BeNil())
		})
	})
})
