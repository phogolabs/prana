package sqlmigr_test

import (
	"fmt"

	"github.com/phogolabs/prana/sqlmigr"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Migration", func() {
	Describe("Parse", func() {
		It("parses the item successfully", func() {
			filename := "20060102150405_schema.sql"
			item, err := sqlmigr.Parse(filename)
			Expect(err).NotTo(HaveOccurred())
			Expect(item.ID).To(Equal("20060102150405"))
			Expect(item.Description).To(Equal("schema"))
			Expect(item.Filenames()).To(ContainElement(filename))
		})

		Context("when the filename is has longer description", func() {
			It("parses the item successfully", func() {
				filename := "20060102150405_my_schema_for_this_db.sql"
				item, err := sqlmigr.Parse(filename)
				Expect(err).NotTo(HaveOccurred())
				Expect(item.ID).To(Equal("20060102150405"))
				Expect(item.Description).To(Equal("my_schema_for_this_db"))
				Expect(item.Drivers).To(ContainElement("sql"))
				Expect(item.Filenames()).To(ContainElement(filename))
			})

			Context("when the filename has driver name as suffix", func() {
				It("parses the item successfully", func() {
					filename := "20060102150405_my_schema_for_this_db_sqlite3.sql"
					item, err := sqlmigr.Parse(filename)
					Expect(err).NotTo(HaveOccurred())
					Expect(item.ID).To(Equal("20060102150405"))
					Expect(item.Description).To(Equal("my_schema_for_this_db"))
					Expect(item.Drivers).To(ContainElement("sqlite3"))
					Expect(item.Filenames()).To(ContainElement(filename))
				})
			})
		})

		Context("when the filename does not contain two parts", func() {
			It("returns an error", func() {
				filename := "schema.sql"
				item, err := sqlmigr.Parse(filename)
				Expect(err).To(MatchError("migration 'schema.sql' has an invalid file name"))
				Expect(item).To(BeNil())
			})
		})

		Context("when the filename does not have timestamp in its name", func() {
			It("returns an error", func() {
				filename := "id_schema.sql"
				item, err := sqlmigr.Parse(filename)
				Expect(err).To(MatchError("migration 'id_schema.sql' has an invalid file name"))
				Expect(item).To(BeNil())
			})
		})
	})
})

var _ = Describe("RunnerErr", func() {
	It("returns the error message", func() {
		err := &sqlmigr.RunnerError{
			Err:       fmt.Errorf("oh no!"),
			Statement: "statement",
		}

		Expect(err).To(MatchError("oh no!: statement"))
	})

	Context("when it has a new line", func() {
		It("returns the error message", func() {
			err := &sqlmigr.RunnerError{
				Err:       fmt.Errorf("oh no!"),
				Statement: "statement\nhello",
			}

			Expect(err).To(MatchError("oh no!: statement"))
		})
	})
})

var _ = Describe("IsNotExist", func() {
	Context("when the error is SQLite error", func() {
		It("returns true", func() {
			err := fmt.Errorf("no such table: migrations")
			Expect(sqlmigr.IsNotExist(err)).To(BeTrue())
		})
	})

	Context("when the error is PostgreSQL error", func() {
		It("returns true", func() {
			err := fmt.Errorf(`pq: relation "migrations" does not exist`)
			Expect(sqlmigr.IsNotExist(err)).To(BeTrue())
		})
	})

	Context("when the error is MySQL error", func() {
		It("returns true", func() {
			err := fmt.Errorf("migrations' doesn't exist")
			Expect(sqlmigr.IsNotExist(err)).To(BeTrue())
		})
	})

	Context("when the error is not supported", func() {
		It("returns false", func() {
			err := fmt.Errorf("oh no")
			Expect(sqlmigr.IsNotExist(err)).To(BeFalse())
		})
	})
})
