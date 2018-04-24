package prana_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/prana"
)

var _ = Describe("ParseURL", func() {
	It("parses the SQLite connection string successfully", func() {
		driver, source, err := prana.ParseURL("sqlite3://./prana.db")
		Expect(err).NotTo(HaveOccurred())
		Expect(driver).To(Equal("sqlite3"))
		Expect(source).To(Equal("./prana.db"))
	})

	It("parses the MySQL connection string successfully", func() {
		driver, source, err := prana.ParseURL("mysql://root@/prana")
		Expect(err).NotTo(HaveOccurred())
		Expect(driver).To(Equal("mysql"))
		Expect(source).To(Equal("root@/prana"))
	})

	It("parses the PostgreSQL connection string successfully", func() {
		driver, source, err := prana.ParseURL("postgres://localhost/prana?sslmode=disable")
		Expect(err).NotTo(HaveOccurred())
		Expect(driver).To(Equal("postgres"))
		Expect(source).To(Equal("postgres://localhost/prana?sslmode=disable"))
	})

	Context("when the URL is invalid", func() {
		It("returns an error", func() {
			driver, source, err := prana.ParseURL("::")
			Expect(driver).To(BeEmpty())
			Expect(source).To(BeEmpty())
			Expect(err).To(MatchError("parse ::: missing protocol scheme"))
		})
	})
})
