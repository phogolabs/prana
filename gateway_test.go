package gom_test

import (
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	lk "github.com/ulule/loukoum"

	_ "github.com/mattn/go-sqlite3"
	"github.com/phogolabs/gom"
	"github.com/phogolabs/gom/script"
)

var _ = Describe("Gateway", func() {
	var db *gom.Gateway

	Describe("Open", func() {
		Context("when cannot open the database", func() {
			It("returns an error", func() {
				db, err := gom.Open("sqlite4", "/tmp/gom.db")
				Expect(db).To(BeNil())
				Expect(err).To(MatchError(`sql: unknown driver "sqlite4" (forgotten import?)`))
			})
		})
	})

	Describe("API", func() {
		type Person struct {
			FirstName string `db:"first_name"`
			LastName  string `db:"last_name"`
			Email     string `db:"email"`
		}

		BeforeEach(func() {
			var err error
			db, err = gom.Open("sqlite3", "/tmp/gom.db")
			Expect(err).To(BeNil())

			buffer := &bytes.Buffer{}
			fmt.Fprintln(buffer, "-- name: create-person-table")
			fmt.Fprintln(buffer, "CREATE TABLE users (")
			fmt.Fprintln(buffer, "first_name text,")
			fmt.Fprintln(buffer, "last_name text,")
			fmt.Fprintln(buffer, "email text")
			fmt.Fprintln(buffer, ");")
			fmt.Fprintln(buffer)

			_, err = db.DB().Exec(buffer.String())
			Expect(err).To(BeNil())

			_, err = db.DB().Exec("INSERT INTO users VALUES('John', 'Doe', 'john@example.com')")
			Expect(err).To(Succeed())
		})

		AfterEach(func() {
			_, err := db.DB().Exec("DROP TABLE users")
			Expect(err).To(BeNil())
			Expect(db.Close()).To(Succeed())
		})

		Describe("Select", func() {
			It("executes a query successfully", func() {
				query := lk.Select("first_name", "last_name", "email").From("users")

				persons := []Person{}
				Expect(db.Select(&persons, query)).To(Succeed())
				Expect(persons).To(HaveLen(1))
				Expect(persons[0].FirstName).To(Equal("John"))
				Expect(persons[0].LastName).To(Equal("Doe"))
				Expect(persons[0].Email).To(Equal("john@example.com"))
			})

			Context("when the query fails", func() {
				It("returns an error", func() {
					query := lk.Select("name").From("categories")

					persons := []Person{}
					Expect(db.Select(&persons, query)).To(MatchError("no such table: categories"))
					Expect(persons).To(BeEmpty())
				})
			})

			Context("when an embedded statement is used", func() {
				It("executes a query successfully", func() {
					query := script.SQL("SELECT * FROM users WHERE first_name = ?", "John")

					persons := []Person{}
					Expect(db.Select(&persons, query)).To(Succeed())
					Expect(persons).To(HaveLen(1))
					Expect(persons[0].FirstName).To(Equal("John"))
					Expect(persons[0].LastName).To(Equal("Doe"))
					Expect(persons[0].Email).To(Equal("john@example.com"))
				})

				Context("when the query does not exist", func() {
					It("returns an error", func() {
						query := script.SQL("SELECT * FROM categories")

						persons := []Person{}
						Expect(db.Select(&persons, query)).To(MatchError("no such table: categories"))
						Expect(persons).To(BeEmpty())
					})
				})
			})
		})

		Describe("SelectOne", func() {
			It("executes a query successfully", func() {
				query := lk.Select("first_name", "last_name", "email").From("users")

				person := Person{}
				Expect(db.SelectOne(&person, query)).To(Succeed())
			})

			Context("when the query fails", func() {
				It("returns an error", func() {
					query := lk.Select("name").From("categories")

					person := Person{}
					Expect(db.SelectOne(&person, query)).To(MatchError("no such table: categories"))
				})
			})
		})

		Describe("Query", func() {
			It("executes a query successfully", func() {
				query := lk.Select("first_name", "last_name", "email").From("users")

				var (
					firstName string
					lastName  string
					email     string
				)

				rows, err := db.Query(query)
				Expect(err).To(BeNil())
				Expect(rows.Next()).To(BeTrue())

				Expect(rows.Scan(&firstName, &lastName, &email)).To(Succeed())
				Expect(firstName).To(Equal("John"))
				Expect(lastName).To(Equal("Doe"))
				Expect(email).To(Equal("john@example.com"))

				Expect(rows.Next()).To(BeFalse())
				Expect(rows.Close()).To(Succeed())
			})

			Context("when the query fails", func() {
				It("returns an error", func() {
					query := lk.Select("name").From("categories")

					rows, err := db.Query(query)
					Expect(err).To(MatchError("no such table: categories"))
					Expect(rows).To(BeNil())
				})
			})
		})

		Describe("QueryRow", func() {
			It("executes a query successfully", func() {
				query := lk.Select("first_name", "last_name", "email").From("users")

				row, err := db.QueryRow(query)
				Expect(err).To(BeNil())
				Expect(row).NotTo(BeNil())

				var (
					firstName string
					lastName  string
					email     string
				)

				Expect(row.Scan(&firstName, &lastName, &email)).To(Succeed())
				Expect(firstName).To(Equal("John"))
				Expect(lastName).To(Equal("Doe"))
				Expect(email).To(Equal("john@example.com"))
			})

			Context("when the query fails", func() {
				It("returns an error", func() {
					query := lk.Select("name").From("categories")

					row, err := db.QueryRow(query)
					Expect(err).To(MatchError("no such table: categories"))
					Expect(row).To(BeNil())
				})
			})
		})

		Describe("Exec", func() {
			It("executes a query successfully", func() {
				query := lk.Delete("users")

				_, err := db.Exec(query)
				Expect(err).To(Succeed())

				rows, err := db.DB().Query("SELECT * FROM users")
				Expect(err).To(BeNil())
				Expect(rows).NotTo(BeNil())
				Expect(rows.Next()).To(BeFalse())
				Expect(rows.Close()).To(Succeed())
			})

			Context("when the query fails", func() {
				It("returns an error", func() {
					query := lk.Delete("categories")
					_, err := db.Exec(query)
					Expect(err).To(MatchError("no such table: categories"))
				})
			})
		})
	})
})
