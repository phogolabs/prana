package gom_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
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
		Expect(gom.LoadSQLCommandsFromReader(buffer)).To(Succeed())
	})

	It("returns a command", func() {
		stmt := gom.Command(script)
		Expect(stmt).NotTo(BeNil())

		query, params := stmt.Prepare()
		Expect(query).To(Equal("SELECT * FROM users"))
		Expect(params).To(BeEmpty())
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

var _ = Describe("Migrate", func() {
	It("passes the migrations to underlying migration executor", func() {
		dir, err := ioutil.TempDir("", "gom_generator")
		Expect(err).To(BeNil())

		url := filepath.Join(dir, "gom.db")
		db, err := gom.Open("sqlite3", url)
		Expect(err).To(BeNil())
		Expect(gom.Migrate(db, gom.Dir(dir))).To(Succeed())
	})
})
