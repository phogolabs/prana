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
})
