package script_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/svett/gom/script"
)

var _ = Describe("Command", func() {
	It("prepares the command correctly", func() {
		stmt := &script.Cmd{
			Query:  "SELECT * FROM users WHERE id = ?",
			Params: []script.Param{1},
		}

		query, params := stmt.Prepare()
		Expect(query).To(Equal("SELECT * FROM users WHERE id = :arg0"))
		Expect(params).To(HaveKeyWithValue("arg0", 1))
	})
})
