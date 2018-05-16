package sqlexec_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/phogolabs/parcello"
	"github.com/phogolabs/prana/fake"
	"github.com/phogolabs/prana/sqlexec"
)

var _ = Describe("Provider", func() {
	var provider *sqlexec.Provider

	BeforeEach(func() {
		provider = &sqlexec.Provider{}
	})

	Describe("ReadFrom", func() {
		var buffer *bytes.Buffer

		BeforeEach(func() {
			buffer = bytes.NewBufferString("-- name: up")
			fmt.Fprintln(buffer)
			fmt.Fprintln(buffer, "SELECT * FROM users;")
		})

		It("loads the provider successfully", func() {
			n, err := provider.ReadFrom(buffer)
			Expect(n).To(Equal(int64(1)))
			Expect(err).To(Succeed())

			query, err := provider.Query("up")
			Expect(err).To(BeNil())
			Expect(query).To(Equal("SELECT * FROM users;"))
		})

		Context("when the statement are duplicated", func() {
			It("returns an error", func() {
				n, err := provider.ReadFrom(buffer)
				Expect(n).To(Equal(int64(1)))
				Expect(err).To(Succeed())

				buffer = bytes.NewBufferString("-- name: up")
				fmt.Fprintln(buffer)
				fmt.Fprintln(buffer, "SELECT * FROM categories;")

				n, err = provider.ReadFrom(buffer)
				Expect(n).To(BeZero())
				Expect(err).To(MatchError("query 'up' already exists"))
			})
		})
	})

	Describe("ReadDir", func() {
		var (
			fileSystem *fake.FileSystem
			buffer     *bytes.Buffer
		)

		BeforeEach(func() {
			fileSystem = &fake.FileSystem{}

			buffer = bytes.NewBufferString("-- name: up")
			fmt.Fprintln(buffer)
			fmt.Fprintln(buffer, "SELECT * FROM categories;")
		})

		It("loads the provider successfully", func() {
			Expect(provider.ReadDir(fileSystem)).To(Succeed())
			Expect(fileSystem.WalkCallCount()).To(Equal(1))

			dir, _ := fileSystem.WalkArgsForCall(0)
			Expect(dir).To(Equal("/"))
		})

		Context("when the driver is provided", func() {
			var node *parcello.Node

			BeforeEach(func() {
				provider.DriverName = "sqlite3"

				data := buffer.Bytes()
				node = &parcello.Node{
					Name:    "file_sqlite3.sql",
					Content: &data,
					Mutex:   &sync.RWMutex{},
				}

				fileSystem.OpenFileReturns(parcello.NewResourceFile(node), nil)
				fileSystem.WalkStub = func(dir string, fn filepath.WalkFunc) error {
					return fn(node.Name, &parcello.ResourceFileInfo{Node: node}, nil)
				}
			})

			It("loads the file successfully", func() {
				Expect(provider.ReadDir(fileSystem)).To(Succeed())
				cmd, err := provider.Query("up")
				Expect(cmd).NotTo(BeNil())
				Expect(err).To(BeNil())
			})

			Context("when the file has another suffix", func() {
				BeforeEach(func() {
					node.Name = "file_sqlite3.sql"
				})

				It("loads the file successfully", func() {
					Expect(provider.ReadDir(fileSystem)).To(Succeed())
					cmd, err := provider.Query("up")
					Expect(cmd).NotTo(BeNil())
					Expect(err).To(BeNil())
				})
			})

			Context("when the driver is not supported by the provider", func() {
				BeforeEach(func() {
					provider.DriverName = "dummy"
				})

				It("does not load the driver", func() {
					Expect(provider.ReadDir(fileSystem)).To(Succeed())
					cmd, err := provider.Query("up")
					Expect(cmd).To(BeEmpty())
					Expect(err).To(MatchError("query 'up' not found"))
				})
			})
		})

		It("skips non sql files", func() {
			data := []byte{}
			node := &parcello.Node{
				Name:    "file.txt",
				Content: &data,
				Mutex:   &sync.RWMutex{},
			}

			fileSystem.OpenFileReturns(parcello.NewResourceFile(node), nil)
			fileSystem.WalkStub = func(dir string, fn filepath.WalkFunc) error {
				return fn(dir, &parcello.ResourceFileInfo{Node: node}, nil)
			}

			Expect(provider.ReadDir(fileSystem)).To(Succeed())

			cmd, err := provider.Query("up")
			Expect(cmd).To(BeEmpty())
			Expect(err).To(MatchError("query 'up' not found"))
		})

		Context("when the file system fails ", func() {
			BeforeEach(func() {
				fileSystem.WalkStub = func(dir string, fn filepath.WalkFunc) error {
					return fn("file.sql", &parcello.ResourceFileInfo{Node: &parcello.Node{Name: "file.sql"}}, nil)
				}
			})

			It("returns an error", func() {
				fileSystem.WalkReturns(fmt.Errorf("Oh no!"))
				Expect(provider.ReadDir(fileSystem)).To(MatchError("Oh no!"))
			})

			Context("when opening a file fails", func() {
				It("returns an error", func() {
					fileSystem.OpenFileReturns(nil, fmt.Errorf("Oh no!"))
					Expect(provider.ReadDir(fileSystem)).To(MatchError("Oh no!"))
				})
			})

			Context("when reading from a file fails", func() {
				It("returns an error", func() {
					data := buffer.Bytes()
					node := &parcello.Node{
						Name:    "file.sql",
						Content: &data,
						Mutex:   &sync.RWMutex{},
					}

					fileSystem.OpenFileReturns(parcello.NewResourceFile(node), nil)
					Expect(provider.ReadDir(fileSystem)).To(Succeed())

					fileSystem.OpenFileReturns(parcello.NewResourceFile(node), nil)
					Expect(provider.ReadDir(fileSystem)).To(MatchError("query 'up' already exists"))
				})
			})
		})
	})

	Describe("Query", func() {
		BeforeEach(func() {
			buffer := bytes.NewBufferString("-- name: up")
			fmt.Fprintln(buffer)
			fmt.Fprintln(buffer, "SELECT * FROM users")

			n, err := provider.ReadFrom(buffer)
			Expect(n).To(Equal(int64(1)))
			Expect(err).To(Succeed())
		})

		It("returns a command", func() {
			query, err := provider.Query("up")
			Expect(err).To(BeNil())
			Expect(query).To(Equal("SELECT * FROM users"))
		})

		Context("when the command has arguments", func() {
			BeforeEach(func() {
				buffer := bytes.NewBufferString("-- name: show-users")
				fmt.Fprintln(buffer)
				fmt.Fprintln(buffer, "SELECT * FROM users WHERE id = ?")

				n, err := provider.ReadFrom(buffer)
				Expect(n).To(Equal(int64(1)))
				Expect(err).To(Succeed())

				provider.DriverName = "ora"
			})

			It("returns a command with params", func() {
				query, err := provider.Query("show-users")
				Expect(err).To(BeNil())
				Expect(query).To(Equal("SELECT * FROM users WHERE id = :arg1"))
			})
		})

		Context("when statements are not found", func() {
			Describe("Cmd", func() {
				It("returns a error", func() {
					cmd, err := provider.Query("down")
					Expect(err).To(MatchError("query 'down' not found"))
					Expect(cmd).To(BeEmpty())
				})
			})
		})
	})
})
