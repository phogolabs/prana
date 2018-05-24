package sqlexec_test

import (
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/prana/sqlexec"
)

var _ = Describe("Splitter", func() {
	var (
		splitter *sqlexec.Splitter
		query    *bytes.Buffer
	)

	BeforeEach(func() {
		query = &bytes.Buffer{}
		fmt.Fprintln(query, "SELECT * FROM users;")
		fmt.Fprintln(query, "GO")
		fmt.Fprintln(query, "SELECT * FROM documents;")

		splitter = &sqlexec.Splitter{}
	})

	ItSplitsTheQuery := func() {
		It("splits query by the separator", func() {
			queries := splitter.Split(query)
			Expect(queries).To(HaveLen(2))
			Expect(queries[0]).To(Equal("SELECT * FROM users;\n"))
			Expect(queries[1]).To(Equal("SELECT * FROM documents;\n"))
		})
	}

	ItSplitsTheQuery()

	Context("when the separator is GO;", func() {
		BeforeEach(func() {
			query.Reset()
			fmt.Fprintln(query, "SELECT * FROM users;")
			fmt.Fprintln(query, "GO;")
			fmt.Fprintln(query, "SELECT * FROM documents;")
		})

		ItSplitsTheQuery()
	})

	Context("when the separator has space", func() {
		BeforeEach(func() {
			query.Reset()
			fmt.Fprintln(query, "SELECT * FROM users;")
			fmt.Fprintln(query, "GO; ")
			fmt.Fprintln(query, "SELECT * FROM documents;")
		})

		ItSplitsTheQuery()
	})

	Context("when the separator has comment", func() {
		BeforeEach(func() {
			query.Reset()
			fmt.Fprintln(query, "SELECT * FROM users;")
			fmt.Fprintln(query, "GO; -- split")
			fmt.Fprintln(query, "SELECT * FROM documents;")
		})

		ItSplitsTheQuery()
	})

	Context("when the separator is in the end", func() {
		BeforeEach(func() {
			query.Reset()
			fmt.Fprintln(query, "SELECT * FROM users;")
			fmt.Fprintln(query, "GO")
		})

		It("remove the GO separator", func() {
			queries := splitter.Split(query)
			Expect(queries).To(HaveLen(1))
			Expect(queries[0]).To(Equal("SELECT * FROM users;\n"))
		})
	})

	Context("when the separator is missing", func() {
		BeforeEach(func() {
			query.Reset()
			fmt.Fprintln(query, "SELECT * FROM users;")
			fmt.Fprintln(query, "SELECT * FROM documents;")
		})

		It("does not split the query", func() {
			stmt := query.String()
			queries := splitter.Split(query)
			Expect(queries).To(HaveLen(1))
			Expect(queries[0]).To(Equal(stmt))
		})
	})
})
