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

	Describe("MySQL", func() {
		It("parses the MySQL connection string successfully", func() {
			driver, source, err := prana.ParseURL("mysql://root@/prana")
			Expect(err).NotTo(HaveOccurred())
			Expect(driver).To(Equal("mysql"))
			Expect(source).To(Equal("root@tcp(127.0.0.1:3306)/prana?parseTime=true"))
		})

		It("parses the MySQL connection string with custom port successfully", func() {
			driver, source, err := prana.ParseURL("mysql://root:password@tcp(127.0.0.1:13306)/prana")
			Expect(err).NotTo(HaveOccurred())
			Expect(driver).To(Equal("mysql"))
			Expect(source).To(Equal("root:password@tcp(127.0.0.1:13306)/prana?parseTime=true"))
		})

		Context("when the DSN is invalid", func() {
			It("returns the error", func() {
				driver, source, err := prana.ParseURL("mysql://@net(addr/")
				Expect(err).To(MatchError("invalid DSN: network address not terminated (missing closing brace)"))
				Expect(driver).To(BeEmpty())
				Expect(source).To(BeEmpty())
			})
		})
	})

	It("parses the PostgreSQL connection string successfully", func() {
		driver, source, err := prana.ParseURL("postgres://localhost/prana?sslmode=disable")
		Expect(err).NotTo(HaveOccurred())
		Expect(driver).To(Equal("postgres"))
		Expect(source).To(Equal("postgres://localhost/prana?sslmode=disable"))
	})

	Context("when the URL is empty", func() {
		It("returns an error", func() {
			driver, source, err := prana.ParseURL("")
			Expect(driver).To(BeEmpty())
			Expect(source).To(BeEmpty())
			Expect(err).To(MatchError("URL cannot be empty"))
		})
	})

	Context("when the URL is invalid", func() {
		It("returns an error", func() {
			driver, source, err := prana.ParseURL("::")
			Expect(driver).To(BeEmpty())
			Expect(source).To(BeEmpty())
			Expect(err).To(MatchError("Invalid DSN"))
		})
	})
})
