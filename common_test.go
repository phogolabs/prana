package gom_test

import (
	"bytes"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/phogolabs/gom"
)

var _ = Describe("Command", func() {
	var script string

	BeforeEach(func() {
		script = fmt.Sprintf("%v", time.Now().UnixNano())
		buffer := bytes.NewBufferString(fmt.Sprintf("-- name: %v", script))
		fmt.Fprintln(buffer)
		fmt.Fprintln(buffer, "SELECT * FROM users")
		Expect(gom.Load(buffer)).To(Succeed())
	})

	It("returns a command", func() {
		stmt := gom.Command(script)
		Expect(stmt).NotTo(BeNil())
		Expect(stmt.Params).To(BeEmpty())
		Expect(stmt.Query).To(Equal("SELECT * FROM users"))
	})

	It("returns a command with params", func() {
		stmt := gom.Command(script, 1)
		Expect(stmt).NotTo(BeNil())
		Expect(stmt.Params).To(ContainElement(1))
		Expect(stmt.Query).To(Equal("SELECT * FROM users"))
	})

	Context("when the statement does not exits", func() {
		It("does not return a statement", func() {
			Expect(func() { gom.Command("down") }).To(Panic())
		})
	})
})

var _ = Describe("ParseURL", func() {
	It("parses the SQLite connection string successfully", func() {
		driver, source, err := gom.ParseURL("sqlite3://./gom.db")
		Expect(err).NotTo(HaveOccurred())
		Expect(driver).To(Equal("sqlite3"))
		Expect(source).To(Equal("./gom.db"))
	})

	It("parses the MySQL connection string successfully", func() {
		driver, source, err := gom.ParseURL("mysql://root@/gom")
		Expect(err).NotTo(HaveOccurred())
		Expect(driver).To(Equal("mysql"))
		Expect(source).To(Equal("root@/gom"))
	})

	It("parses the PostgreSQL connection string successfully", func() {
		driver, source, err := gom.ParseURL("postgres://localhost/gom?sslmode=disable")
		Expect(err).NotTo(HaveOccurred())
		Expect(driver).To(Equal("postgres"))
		Expect(source).To(Equal("postgres://localhost/gom?sslmode=disable"))
	})

	Context("when the URL is invalid", func() {
		It("returns an error", func() {
			driver, source, err := gom.ParseURL("::")
			Expect(driver).To(BeEmpty())
			Expect(source).To(BeEmpty())
			Expect(err).To(MatchError("parse ::: missing protocol scheme"))
		})
	})
})
