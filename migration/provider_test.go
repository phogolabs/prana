package migration_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/gom"
	"github.com/phogolabs/gom/migration"
)

var _ = Describe("Provider", func() {
	var provider *migration.Provider

	BeforeEach(func() {
		dir, err := ioutil.TempDir("", "gom_runner")
		Expect(err).To(BeNil())

		db := filepath.Join(dir, "gom.db")
		gateway, err := gom.Open("sqlite3", db)
		Expect(err).To(BeNil())

		provider = &migration.Provider{
			Dir:     dir,
			Gateway: gateway,
		}
	})

	JustBeforeEach(func() {
		query := &bytes.Buffer{}
		fmt.Fprintln(query, "CREATE TABLE migrations (")
		fmt.Fprintln(query, " id          TEXT      NOT NULL PRIMARY KEY,")
		fmt.Fprintln(query, " description TEXT      NOT NULL,")
		fmt.Fprintln(query, " created_at  TIMESTAMP NOT NULL")
		fmt.Fprintln(query, ");")

		_, err := provider.Gateway.DB().Exec(query.String())
		Expect(err).To(BeNil())

		path := filepath.Join(provider.Dir, "20060102150405_schema.sql")
		Expect(ioutil.WriteFile(path, []byte{}, 0700)).To(Succeed())

		insert := gom.Insert("migrations").Set(
			gom.Pair("id", "20060102150405"),
			gom.Pair("description", "schema"),
			gom.Pair("created_at", time.Now()),
		)

		_, err = provider.Gateway.Exec(insert)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		provider.Gateway.Close()
	})

	Describe("Insert", func() {
		It("inserts a migration item successfully", func() {
			item := migration.Item{
				Id:          "20070102150405",
				Description: "trigger",
			}

			Expect(provider.Insert(&item)).To(Succeed())

			items := []migration.Item{}
			query := gom.Select().From("migrations").OrderBy(gom.Order("id", gom.Asc))

			Expect(provider.Gateway.Select(&items, query)).To(Succeed())
			Expect(items).To(HaveLen(2))

			Expect(items[0].Id).To(Equal("20060102150405"))
			Expect(items[0].Description).To(Equal("schema"))

			Expect(items[1].Id).To(Equal("20070102150405"))
			Expect(items[1].Description).To(Equal("trigger"))
		})

		Context("when the database is not available", func() {
			JustBeforeEach(func() {
				Expect(provider.Gateway.Close()).To(Succeed())
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
			query := gom.Select().From("migrations")

			Expect(provider.Gateway.Select(&items, query)).To(Succeed())
			Expect(items).To(BeEmpty())
		})

		Context("when the database is not available", func() {
			JustBeforeEach(func() {
				Expect(provider.Gateway.Close()).To(Succeed())
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
			path := filepath.Join(provider.Dir, "20070102150405_setup.sql")
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
				path := provider.Dir + "_old"
				Expect(os.Rename(provider.Dir, path)).To(Succeed())
			})

			It("returns an error", func() {
				msg := fmt.Sprintf("Directory '%s' does not exist", provider.Dir)
				items, err := provider.Migrations()
				Expect(items).To(BeEmpty())
				Expect(err).To(MatchError(msg))
			})
		})

		Context("when cannot parse a migration", func() {
			JustBeforeEach(func() {
				old := filepath.Join(provider.Dir, "20060102150405_schema.sql")
				new := filepath.Join(provider.Dir, "id_schema.sql")
				Expect(os.Rename(old, new)).To(Succeed())
			})

			It("returns an error", func() {
				path := filepath.Join(provider.Dir, "id_schema.sql")
				msg := fmt.Sprintf("Migration '%s' has an invalid file name", path)

				items, err := provider.Migrations()
				Expect(items).To(BeEmpty())
				Expect(err).To(MatchError(msg))
			})
		})

		Context("when the database is not available", func() {
			JustBeforeEach(func() {
				Expect(provider.Gateway.Close()).To(Succeed())
			})

			It("returns an error", func() {
				items, err := provider.Migrations()
				Expect(items).To(BeEmpty())
				Expect(err).To(MatchError("sql: database is closed"))
			})
		})

		Context("when the migration has ID mismatch", func() {
			JustBeforeEach(func() {
				old := filepath.Join(provider.Dir, "20060102150405_schema.sql")
				new := filepath.Join(provider.Dir, "20070102150405_schema.sql")
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
				old := filepath.Join(provider.Dir, "20060102150405_schema.sql")
				new := filepath.Join(provider.Dir, "20060102150405_tables.sql")
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
