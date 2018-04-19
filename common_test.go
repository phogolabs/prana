package oak_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/phogolabs/oak"
	"github.com/phogolabs/oak/fake"
	"github.com/phogolabs/parcello"
)

var _ = Describe("Command", func() {
	var script string

	BeforeEach(func() {
		script = fmt.Sprintf("%v", time.Now().UnixNano())
		buffer := bytes.NewBufferString(fmt.Sprintf("-- name: %v", script))
		fmt.Fprintln(buffer)
		fmt.Fprintln(buffer, "SELECT * FROM users")
		Expect(oak.LoadSQLCommandsFromReader(buffer)).To(Succeed())

	})

	It("returns a command", func() {
		stmt := oak.Command(script)
		Expect(stmt).NotTo(BeNil())

		query, params := stmt.Prepare()
		Expect(query).To(Equal("SELECT * FROM users"))
		Expect(params).To(BeEmpty())
	})

	Context("when loading a whole directory", func() {
		BeforeEach(func() {
			buffer := bytes.NewBufferString(fmt.Sprintf("-- name: %v", "cmd"))
			fmt.Fprintln(buffer)
			fmt.Fprintln(buffer, "SELECT * FROM categories")

			content := buffer.Bytes()

			node := &parcello.Node{
				Name:    "script.sql",
				Content: &content,
				Mutex:   &sync.RWMutex{},
			}

			fileSystem := &fake.FileSystem{}
			fileSystem.OpenFileReturns(parcello.NewResourceFile(node), nil)

			fileSystem.WalkStub = func(dir string, fn filepath.WalkFunc) error {
				return fn("command", &parcello.ResourceFileInfo{Node: node}, nil)
			}

			Expect(oak.LoadSQLCommandsFrom(fileSystem)).To(Succeed())
		})

		It("returns a command", func() {
			stmt := oak.Command("cmd")
			Expect(stmt).NotTo(BeNil())

			query, params := stmt.Prepare()
			Expect(query).To(Equal("SELECT * FROM categories"))
			Expect(params).To(BeEmpty())
		})
	})

	Context("when the statement does not exits", func() {
		It("does not return a statement", func() {
			Expect(func() { oak.Command("down") }).To(Panic())
		})
	})
})

var _ = Describe("ParseURL", func() {
	It("parses the SQLite connection string successfully", func() {
		driver, source, err := oak.ParseURL("sqlite3://./oak.db")
		Expect(err).NotTo(HaveOccurred())
		Expect(driver).To(Equal("sqlite3"))
		Expect(source).To(Equal("./oak.db"))
	})

	It("parses the MySQL connection string successfully", func() {
		driver, source, err := oak.ParseURL("mysql://root@/oak")
		Expect(err).NotTo(HaveOccurred())
		Expect(driver).To(Equal("mysql"))
		Expect(source).To(Equal("root@/oak"))
	})

	It("parses the PostgreSQL connection string successfully", func() {
		driver, source, err := oak.ParseURL("postgres://localhost/oak?sslmode=disable")
		Expect(err).NotTo(HaveOccurred())
		Expect(driver).To(Equal("postgres"))
		Expect(source).To(Equal("postgres://localhost/oak?sslmode=disable"))
	})

	Context("when the URL is invalid", func() {
		It("returns an error", func() {
			driver, source, err := oak.ParseURL("::")
			Expect(driver).To(BeEmpty())
			Expect(source).To(BeEmpty())
			Expect(err).To(MatchError("parse ::: missing protocol scheme"))
		})
	})
})

var _ = Describe("Migrate", func() {
	It("passes the migrations to underlying migration executor", func() {
		dir, err := ioutil.TempDir("", "oak_generator")
		Expect(err).To(BeNil())

		url := filepath.Join(dir, "oak.db")
		db, err := oak.Open("sqlite3", url)
		Expect(err).To(BeNil())
		Expect(oak.Migrate(db, parcello.Dir(dir))).To(Succeed())
	})
})
