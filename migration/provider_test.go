package migration_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/gom"
	"github.com/phogolabs/gom/migration"
)

var _ = Describe("Provider", func() {
	var (
		provider *migration.Provider
		dir      string
	)

	BeforeEach(func() {
		var err error

		dir, err = ioutil.TempDir("", "gom_runner")
		Expect(err).To(BeNil())

		conn := filepath.Join(dir, "gom.db")
		db, err := sqlx.Open("sqlite3", conn)
		Expect(err).To(BeNil())

		provider = &migration.Provider{
			Dir:        "/",
			FileSystem: gom.Dir(dir),
			DB:         db,
		}
	})

	JustBeforeEach(func() {
		query := &bytes.Buffer{}
		fmt.Fprintln(query, "CREATE TABLE migrations (")
		fmt.Fprintln(query, " id          TEXT      NOT NULL PRIMARY KEY,")
		fmt.Fprintln(query, " description TEXT      NOT NULL,")
		fmt.Fprintln(query, " created_at  TIMESTAMP NOT NULL")
		fmt.Fprintln(query, ");")

		_, err := provider.DB.Exec(query.String())
		Expect(err).To(BeNil())

		path := filepath.Join(dir, "20060102150405_schema.sql")
		Expect(ioutil.WriteFile(path, []byte{}, 0700)).To(Succeed())

		insert := "INSERT INTO migrations(id, description, created_at) VALUES(?,?,?)"
		_, err = provider.DB.Exec(insert, "20060102150405", "schema", time.Now())
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		provider.DB.Close()
	})

	Describe("Insert", func() {
		It("inserts a migration item successfully", func() {
			item := migration.Item{
				Id:          "20070102150405",
				Description: "trigger",
			}

			Expect(provider.Insert(&item)).To(Succeed())

			items := []migration.Item{}
			query := "SELECT * FROM migrations ORDER BY id ASC"

			Expect(provider.DB.Select(&items, query)).To(Succeed())
			Expect(items).To(HaveLen(2))

			Expect(items[0].Id).To(Equal("20060102150405"))
			Expect(items[0].Description).To(Equal("schema"))

			Expect(items[1].Id).To(Equal("20070102150405"))
			Expect(items[1].Description).To(Equal("trigger"))
		})

		Context("when the database is not available", func() {
			JustBeforeEach(func() {
				Expect(provider.DB.Close()).To(Succeed())
			})

			It("returns an error", func() {
				items, err := provider.Migrations()
				Expect(items).To(BeEmpty())
				Expect(err).To(MatchError("sql: database is closed"))
			})
		})
	})

	Describe("Delete", func() {
		It("deletes a migration item successfully", func() {
			item := migration.Item{
				Id:          "20060102150405",
				Description: "schema",
			}

			Expect(provider.Delete(&item)).To(Succeed())

			items := []migration.Item{}
			query := "SELECT * FROM migrations"

			Expect(provider.DB.Select(&items, query)).To(Succeed())
			Expect(items).To(BeEmpty())
		})

		Context("when the database is not available", func() {
			JustBeforeEach(func() {
				Expect(provider.DB.Close()).To(Succeed())
			})

			It("returns an error", func() {
				item := migration.Item{
					Id:          "20060102150405",
					Description: "setup",
				}
				Expect(provider.Delete(&item)).To(MatchError("sql: database is closed"))
			})
		})
	})

	Describe("Migrations", func() {
		It("returns the migrations successfully", func() {
			path := filepath.Join(dir, "20070102150405_setup.sql")
			Expect(ioutil.WriteFile(path, []byte{}, 0700)).To(Succeed())

			items, err := provider.Migrations()
			Expect(err).NotTo(HaveOccurred())
			Expect(items).To(HaveLen(2))

			Expect(items[0].Id).To(Equal("20060102150405"))
			Expect(items[0].Description).To(Equal("schema"))

			Expect(items[1].Id).To(Equal("20070102150405"))
			Expect(items[1].Description).To(Equal("setup"))
			Expect(items[1].CreatedAt.IsZero()).To(BeTrue())
		})

		Context("when the directory does not exist", func() {
			JustBeforeEach(func() {
				path := dir + "_old"
				Expect(os.Rename(dir, path)).To(Succeed())
			})

			It("returns an error", func() {
				items, err := provider.Migrations()
				Expect(items).To(BeEmpty())
				Expect(err).To(MatchError("Directory '.' does not exist"))
			})
		})

		Context("when cannot parse a migration", func() {
			JustBeforeEach(func() {
				old := filepath.Join(dir, "20060102150405_schema.sql")
				new := filepath.Join(dir, "id_schema.sql")
				Expect(os.Rename(old, new)).To(Succeed())
			})

			It("returns an error", func() {
				items, err := provider.Migrations()
				Expect(items).To(BeEmpty())
				Expect(err).To(MatchError("Migration 'id_schema.sql' has an invalid file name"))
			})
		})

		Context("when the database is not available", func() {
			JustBeforeEach(func() {
				Expect(provider.DB.Close()).To(Succeed())
			})

			It("returns an error", func() {
				items, err := provider.Migrations()
				Expect(items).To(BeEmpty())
				Expect(err).To(MatchError("sql: database is closed"))
			})
		})

		Context("when the migration has ID mismatch", func() {
			JustBeforeEach(func() {
				old := filepath.Join(dir, "20060102150405_schema.sql")
				new := filepath.Join(dir, "20070102150405_schema.sql")
				Expect(os.Rename(old, new)).To(Succeed())
			})

			It("returns an error", func() {
				items, err := provider.Migrations()
				Expect(items).To(BeEmpty())
				Expect(err).To(MatchError("Mismatched migration id. Expected: '20060102150405' but has '20070102150405'"))
			})
		})

		Context("when the migration has Description mismatch", func() {
			JustBeforeEach(func() {
				old := filepath.Join(dir, "20060102150405_schema.sql")
				new := filepath.Join(dir, "20060102150405_tables.sql")
				Expect(os.Rename(old, new)).To(Succeed())
			})

			It("returns an error", func() {
				items, err := provider.Migrations()
				Expect(items).To(BeEmpty())
				Expect(err).To(MatchError("Mismatched migration description. Expected: 'schema' but has 'tables'"))
			})
		})
	})
})
