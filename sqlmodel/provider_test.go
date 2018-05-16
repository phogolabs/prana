package sqlmodel_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/phogolabs/prana/fake"
	"github.com/phogolabs/prana/sqlmodel"
)

var _ = Describe("PostgreSQLProvider", func() {
	var (
		provider *sqlmodel.PostgreSQLProvider
		db       *sqlx.DB
	)

	BeforeEach(func() {
		var err error

		db, err = sqlx.Open("postgres", "postgres://localhost/prana?sslmode=disable")
		Expect(err).NotTo(HaveOccurred())

		provider = &sqlmodel.PostgreSQLProvider{
			DB: db,
		}
	})

	AfterEach(func() {
		Expect(provider.Close()).To(Succeed())
	})

	Describe("Tables", func() {
		BeforeEach(func() {
			_, err := db.Exec("CREATE TABLE my_table(id serial)")
			Expect(err).NotTo(HaveOccurred())

			_, err = db.Exec("CREATE TABLE your_table(id serial)")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_, err := db.Exec("DROP TABLE IF EXISTS my_table")
			Expect(err).NotTo(HaveOccurred())

			_, err = db.Exec("DROP TABLE IF EXISTS your_table")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the table names successfully", func() {
			tables, err := provider.Tables("public")
			Expect(err).NotTo(HaveOccurred())
			Expect(tables).To(HaveLen(2))
			Expect(tables).To(ContainElement("my_table"))
			Expect(tables).To(ContainElement("your_table"))
		})

		Context("when the schema is empty", func() {
			It("returns the table names for public schema successfully", func() {
				tables, err := provider.Tables("")
				Expect(err).NotTo(HaveOccurred())
				Expect(tables).To(HaveLen(2))
				Expect(tables).To(ContainElement("my_table"))
				Expect(tables).To(ContainElement("your_table"))
			})
		})

		Context("when the database is not available", func() {
			BeforeEach(func() {
				db, err := sqlx.Open("postgres", "postgres://localhost/prana?sslmode=disable")
				Expect(err).NotTo(HaveOccurred())
				Expect(db.Close()).To(Succeed())
				provider.DB = db
			})

			It("return an error", func() {
				tables, err := provider.Tables("public")
				Expect(tables).To(BeEmpty())
				Expect(err).To(MatchError("sql: database is closed"))
			})
		})
	})

	Describe("Schema", func() {
		BeforeEach(func() {
			_, ferr := db.Exec("CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy');")
			Expect(ferr).NotTo(HaveOccurred())

			query := &bytes.Buffer{}

			fmt.Fprintln(query, " varbit_field_null                    varbit(20) NULL,")
			fmt.Fprintln(query, " varbit_field_not_null                varbit(20) NOT NULL,")
			fmt.Fprintln(query, " bit_varying_field_null               bit varying(20) NULL,")
			fmt.Fprintln(query, " bit_varying_field_not_null           bit varying(20) NOT NULL,")
			fmt.Fprintln(query, " smallserial_field_not_null           smallserial NOT NULL,")
			fmt.Fprintln(query, " bigserial_field_not_null             bigserial NOT NULL,")
			fmt.Fprintln(query, " money_field_null                     money NULL,")
			fmt.Fprintln(query, " money_field_not_null                 money NOT NULL,")
			fmt.Fprintln(query, " timestamp_field_not_null             timestamp NOT NULL,")
			fmt.Fprintln(query, " timestamp_without_tz_field_null      timestamp without time zone NULL,")
			fmt.Fprintln(query, " timestamp_without_tz_field_not_null  timestamp without time zone NOT NULL,")
			fmt.Fprintln(query, " timestamp_with_tz_field_null         timestamp with time zone NULL,")
			fmt.Fprintln(query, " timestamp_with_tz_field_not_null     timestamp with time zone NOT NULL,")
			fmt.Fprintln(query, " time_without_tz_field_null           time without time zone NULL,")
			fmt.Fprintln(query, " time_without_tz_field_not_null       time without time zone NOT NULL,")
			fmt.Fprintln(query, " time_with_tz_field_null              time with time zone NULL,")
			fmt.Fprintln(query, " time_with_tz_field_not_null          time with time zone NOT NULL,")
			fmt.Fprintln(query, " bytea_field_null                     bytea NULL,")
			fmt.Fprintln(query, " bytea_field_not_null                 bytea NOT NULL,")
			fmt.Fprintln(query, " jsonb_field_null                     jsonb NULL,")
			fmt.Fprintln(query, " jsonb_field_not_null                 jsonb NOT NULL,")
			fmt.Fprintln(query, " xml_field_null                       xml NULL,")
			fmt.Fprintln(query, " xml_field_not_null                   xml NOT NULL,")
			fmt.Fprintln(query, " uuid_field_null                      uuid NULL,")
			fmt.Fprintln(query, " uuid_field_not_null                  uuid NOT NULL,")
			fmt.Fprintln(query, " hstore_field_null                    hstore NULL,")
			fmt.Fprintln(query, " hstore_field_not_null                hstore NOT NULL,")
			fmt.Fprintln(query, " mood_field_null                      mood NULL,")
			fmt.Fprintln(query, " mood_field_not_null                  mood NOT NULL")

			_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
			Expect(err).NotTo(HaveOccurred())

			_, err = db.Exec("CREATE EXTENSION IF NOT EXISTS \"hstore\"")
			Expect(err).NotTo(HaveOccurred())

			_, err = db.Exec(CreateTable(query))
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {

			_, err := db.Exec("DROP TABLE IF EXISTS test")
			Expect(err).NotTo(HaveOccurred())

			_, err = db.Exec("DROP TYPE mood")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the schema successfully", func() {
			schema, err := provider.Schema("public", "test")
			Expect(err).NotTo(HaveOccurred())
			Expect(schema).NotTo(BeNil())
			Expect(schema.Name).To(Equal("public"))
			Expect(schema.Tables).To(HaveLen(1))

			table := schema.Tables[0]
			Expect(table.Name).To(Equal("test"))
			Expect(table.Columns).To(HaveLen(65))
			ExpectColumnsForPostgreSQL(table.Columns)
		})

		Context("when the table has primary key", func() {
			BeforeEach(func() {
				_, err := db.Exec("CREATE TABLE my_table(id serial primary key)")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				_, err := db.Exec("DROP TABLE IF EXISTS my_table")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the schema successfully", func() {
				schema, err := provider.Schema("", "my_table")
				Expect(err).NotTo(HaveOccurred())
				Expect(schema).NotTo(BeNil())
				Expect(schema.Name).To(Equal("public"))
				Expect(schema.Tables).To(HaveLen(1))

				table := schema.Tables[0]
				Expect(table.Name).To(Equal("my_table"))
				Expect(table.Columns).To(HaveLen(1))

				column := table.Columns[0]
				Expect(column.Name).To(Equal("id"))
				Expect(column.Type.IsPrimaryKey).To(BeTrue())
			})
		})

		Context("when the table names are not provided", func() {
			It("return an error", func() {
				schema, err := provider.Schema("public")
				Expect(schema).To(BeNil())
				Expect(err).To(MatchError("No tables found"))
			})
		})

		Context("when the database is not available", func() {
			BeforeEach(func() {
				db, err := sqlx.Open("postgres", "postgres://localhost/prana?sslmode=disable")
				Expect(err).NotTo(HaveOccurred())
				Expect(db.Close()).To(Succeed())
				provider.DB = db
			})

			It("return an error", func() {
				schema, err := provider.Schema("public", "test")
				Expect(schema).To(BeNil())
				Expect(err).To(MatchError("sql: database is closed"))
			})
		})

		Context("when the table information cannot be fetched", func() {
			BeforeEach(func() {
				querier := &fake.Querier{}
				querier.CloseStub = db.Close
				querier.QueryRowStub = db.QueryRow
				querier.QueryStub = func(txt string, args ...interface{}) (*sql.Rows, error) {
					if strings.Contains(txt, "information_schema.columns") {
						return nil, fmt.Errorf("oh no!")
					}
					return db.Query(txt, args...)
				}

				provider.DB = querier
			})

			It("return an error", func() {
				schema, err := provider.Schema("public", "test")
				Expect(schema).To(BeNil())
				Expect(err).To(MatchError("oh no!"))
			})
		})
	})
})

var _ = Describe("MySQLProvider", func() {
	var (
		provider *sqlmodel.MySQLProvider
		db       *sqlx.DB
	)

	BeforeEach(func() {
		var err error

		db, err = sqlx.Open("mysql", "root@/prana")
		Expect(err).NotTo(HaveOccurred())

		provider = &sqlmodel.MySQLProvider{
			DB: db,
		}
	})

	AfterEach(func() {
		Expect(provider.Close()).To(Succeed())
	})

	Describe("Tables", func() {
		BeforeEach(func() {
			_, err := db.Exec("CREATE TABLE my_table(id serial)")
			Expect(err).NotTo(HaveOccurred())

			_, err = db.Exec("CREATE TABLE your_table(id serial)")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_, err := db.Exec("DROP TABLE IF EXISTS my_table")
			Expect(err).NotTo(HaveOccurred())

			_, err = db.Exec("DROP TABLE IF EXISTS your_table")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the table names successfully", func() {
			tables, err := provider.Tables("")
			Expect(err).NotTo(HaveOccurred())
			Expect(tables).To(HaveLen(2))
			Expect(tables).To(ContainElement("my_table"))
			Expect(tables).To(ContainElement("your_table"))
		})

		Context("when the schema is empty", func() {
			It("returns the table names for public schema successfully", func() {
				tables, err := provider.Tables("")
				Expect(err).NotTo(HaveOccurred())
				Expect(tables).To(HaveLen(2))
				Expect(tables).To(ContainElement("my_table"))
				Expect(tables).To(ContainElement("your_table"))
			})
		})

		Context("when the database is not available", func() {
			BeforeEach(func() {
				dbb, err := sqlx.Open("mysql", "root@/prana")
				Expect(err).NotTo(HaveOccurred())
				Expect(dbb.Close()).To(Succeed())
				provider.DB = dbb
			})

			It("return an error", func() {
				tables, err := provider.Tables("public")
				Expect(tables).To(BeEmpty())
				Expect(err).To(MatchError("sql: database is closed"))
			})

			Context("when the schema is empty", func() {
				It("return an error", func() {
					tables, err := provider.Tables("")
					Expect(tables).To(BeEmpty())
					Expect(err).To(MatchError("sql: database is closed"))
				})
			})
		})
	})

	Describe("Schema", func() {
		BeforeEach(func() {
			query := &bytes.Buffer{}

			fmt.Fprintln(query, " bit_tinyint_field_unsigned_null       tinyint(1) UNSIGNED NULL,")
			fmt.Fprintln(query, " bit_tinyint_field_unsigned_not_null   tinyint(1) UNSIGNED NOT NULL,")
			fmt.Fprintln(query, " bit_tinyint_field_null                tinyint(1) NULL,")
			fmt.Fprintln(query, " bit_tinyint_field_not_null            tinyint(1) NOT NULL,")
			fmt.Fprintln(query, " tinyint_field_unsigned_null           tinyint(2) UNSIGNED NULL,")
			fmt.Fprintln(query, " tinyint_field_unsigned_not_null       tinyint(2) UNSIGNED NOT NULL,")
			fmt.Fprintln(query, " tinyint_field_null                    tinyint(2) NULL,")
			fmt.Fprintln(query, " tinyint_field_not_null                tinyint(2) NOT NULL,")
			fmt.Fprintln(query, " smallint_field_unsigned_null          smallint  UNSIGNED NULL,")
			fmt.Fprintln(query, " smallint_field_unsigned_not_null      smallint  UNSIGNED NOT NULL,")
			fmt.Fprintln(query, " mediumint_field_unsigned_null         mediumint  UNSIGNED NULL,")
			fmt.Fprintln(query, " mediumint_field_unsigned_not_null     mediumint  UNSIGNED NOT NULL,")
			fmt.Fprintln(query, " mediumint_field_null                  mediumint  NULL,")
			fmt.Fprintln(query, " mediumint_field_not_null              mediumint  NOT NULL,")
			fmt.Fprintln(query, " int_field_unsigned_null               int UNSIGNED NULL,")
			fmt.Fprintln(query, " int_field_unsigned_not_null           int UNSIGNED NOT NULL,")
			fmt.Fprintln(query, " varbinary_field_null                  varbinary(20) NULL,")
			fmt.Fprintln(query, " varbinary_field_not_null              varbinary(20) NOT NULL")

			_, err := db.Exec(CreateTable(query))
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_, err := db.Exec("DROP TABLE IF EXISTS test")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the schema successfully", func() {
			schema, err := provider.Schema("", "test")
			Expect(err).NotTo(HaveOccurred())
			Expect(schema).NotTo(BeNil())
			Expect(schema.Name).To(Equal("prana"))
			Expect(schema.Tables).To(HaveLen(1))

			table := schema.Tables[0]
			Expect(table.Name).To(Equal("test"))
			Expect(table.Columns).To(HaveLen(54))
			ExpectColumnsForMySQL(table.Columns)
		})

		Context("when the table has primary key", func() {
			BeforeEach(func() {
				_, err := db.Exec("CREATE TABLE my_table(id serial primary key)")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				_, err := db.Exec("DROP TABLE IF EXISTS my_table")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the schema successfully", func() {
				schema, err := provider.Schema("", "my_table")
				Expect(err).NotTo(HaveOccurred())
				Expect(schema).NotTo(BeNil())
				Expect(schema.Name).To(Equal("prana"))
				Expect(schema.Tables).To(HaveLen(1))

				table := schema.Tables[0]
				Expect(table.Name).To(Equal("my_table"))
				Expect(table.Columns).To(HaveLen(1))

				column := table.Columns[0]
				Expect(column.Name).To(Equal("id"))
				Expect(column.Type.IsPrimaryKey).To(BeTrue())
			})
		})

		Context("when the table names are not provided", func() {
			It("return an error", func() {
				schema, err := provider.Schema("")
				Expect(schema).To(BeNil())
				Expect(err).To(MatchError("No tables found"))
			})
		})

		Context("when the database is not available", func() {
			BeforeEach(func() {
				db, err := sqlx.Open("mysql", "root@/prana")
				Expect(err).NotTo(HaveOccurred())
				Expect(db.Close()).To(Succeed())
				provider.DB = db
			})

			It("return an error", func() {
				schema, err := provider.Schema("", "test")
				Expect(schema).To(BeNil())
				Expect(err).To(MatchError("sql: database is closed"))
			})
		})

		Context("when the table primary key information cannot be fetched", func() {
			BeforeEach(func() {
				querier := &fake.Querier{}
				querier.CloseStub = db.Close
				querier.QueryRowStub = db.QueryRow
				querier.QueryStub = func(txt string, args ...interface{}) (*sql.Rows, error) {
					if strings.Contains(txt, "information_schema.table_constraints") {
						return nil, fmt.Errorf("oh no!")
					}
					return db.Query(txt, args...)
				}

				provider.DB = querier
			})

			It("return an error", func() {
				schema, err := provider.Schema("public", "test")
				Expect(schema).To(BeNil())
				Expect(err).To(MatchError("oh no!"))
			})
		})

		Context("when the table information cannot be fetched", func() {
			BeforeEach(func() {
				querier := &fake.Querier{}
				querier.CloseStub = db.Close
				querier.QueryRowStub = db.QueryRow
				querier.QueryStub = func(txt string, args ...interface{}) (*sql.Rows, error) {
					if strings.Contains(txt, "information_schema.columns") {
						return nil, fmt.Errorf("oh no!")
					}
					return db.Query(txt, args...)
				}

				provider.DB = querier
			})

			It("return an error", func() {
				schema, err := provider.Schema("public", "test")
				Expect(schema).To(BeNil())
				Expect(err).To(MatchError("oh no!"))
			})
		})
	})
})

var _ = Describe("SQLiteProvider", func() {
	var (
		provider *sqlmodel.SQLiteProvider
		db       *sqlx.DB
	)

	BeforeEach(func() {
		var err error

		dir, err := ioutil.TempDir("", "prana")
		Expect(err).To(BeNil())

		conn := filepath.Join(dir, "prana.db")
		db, err = sqlx.Open("sqlite3", conn)
		Expect(err).NotTo(HaveOccurred())

		provider = &sqlmodel.SQLiteProvider{
			DB: db,
		}
	})

	AfterEach(func() {
		Expect(provider.Close()).To(Succeed())
	})

	Describe("Tables", func() {
		BeforeEach(func() {
			_, err := db.Exec("CREATE TABLE my_table(id serial)")
			Expect(err).NotTo(HaveOccurred())

			_, err = db.Exec("CREATE TABLE your_table(id serial)")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_, err := db.Exec("DROP TABLE IF EXISTS my_table")
			Expect(err).NotTo(HaveOccurred())

			_, err = db.Exec("DROP TABLE IF EXISTS your_table")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the table names successfully", func() {
			tables, err := provider.Tables("public")
			Expect(err).NotTo(HaveOccurred())
			Expect(tables).To(HaveLen(2))
			Expect(tables).To(ContainElement("my_table"))
			Expect(tables).To(ContainElement("your_table"))
		})

		Context("when the schema is empty", func() {
			It("returns the table names for public schema successfully", func() {
				tables, err := provider.Tables("")
				Expect(err).NotTo(HaveOccurred())
				Expect(tables).To(HaveLen(2))
				Expect(tables).To(ContainElement("my_table"))
				Expect(tables).To(ContainElement("your_table"))
			})
		})

		Context("when the database is not available", func() {
			BeforeEach(func() {
				dir, err := ioutil.TempDir("", "prana")
				Expect(err).To(BeNil())

				conn := filepath.Join(dir, "prana.db")
				db, err := sqlx.Open("sqlite3", conn)
				Expect(err).NotTo(HaveOccurred())
				Expect(db.Close()).To(Succeed())

				provider.DB = db
			})

			It("return an error", func() {
				tables, err := provider.Tables("public")
				Expect(tables).To(BeEmpty())
				Expect(err).To(MatchError("sql: database is closed"))
			})
		})
	})

	Describe("Schema", func() {
		BeforeEach(func() {
			query := &bytes.Buffer{}

			fmt.Fprintln(query, " varbit_field_null                    varbit(20) NULL,")
			fmt.Fprintln(query, " varbit_field_not_null                varbit(20) NOT NULL,")
			fmt.Fprintln(query, " bit_varying_field_null               bit varying(20) NULL,")
			fmt.Fprintln(query, " bit_varying_field_not_null           bit varying(20) NOT NULL,")
			fmt.Fprintln(query, " smallserial_field_not_null           smallserial NOT NULL,")
			fmt.Fprintln(query, " bigserial_field_not_null             bigserial NOT NULL,")
			fmt.Fprintln(query, " money_field_null                     money NULL,")
			fmt.Fprintln(query, " money_field_not_null                 money NOT NULL,")
			fmt.Fprintln(query, " timestamp_without_tz_field_null      timestamp without time zone NULL,")
			fmt.Fprintln(query, " timestamp_without_tz_field_not_null  timestamp without time zone NOT NULL,")
			fmt.Fprintln(query, " timestamp_with_tz_field_null         timestamp with time zone NULL,")
			fmt.Fprintln(query, " timestamp_with_tz_field_not_null     timestamp with time zone NOT NULL,")
			fmt.Fprintln(query, " time_without_tz_field_null           time without time zone NULL,")
			fmt.Fprintln(query, " time_without_tz_field_not_null       time without time zone NOT NULL,")
			fmt.Fprintln(query, " time_with_tz_field_null              time with time zone NULL,")
			fmt.Fprintln(query, " time_with_tz_field_not_null          time with time zone NOT NULL,")
			fmt.Fprintln(query, " bytea_field_null                     bytea NULL,")
			fmt.Fprintln(query, " bytea_field_not_null                 bytea NOT NULL,")
			fmt.Fprintln(query, " jsonb_field_null                     jsonb NULL,")
			fmt.Fprintln(query, " jsonb_field_not_null                 jsonb NOT NULL,")
			fmt.Fprintln(query, " xml_field_null                       xml NULL,")
			fmt.Fprintln(query, " xml_field_not_null                   xml NOT NULL,")
			fmt.Fprintln(query, " uuid_field_null                      uuid NULL,")
			fmt.Fprintln(query, " uuid_field_not_null                  uuid NOT NULL,")
			fmt.Fprintln(query, " timestamp_field_not_null             timestamp NOT NULL")

			_, err := db.Exec(CreateTable(query))
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_, err := db.Exec("DROP TABLE IF EXISTS test")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the schema successfully", func() {
			schema, err := provider.Schema("", "test")
			Expect(err).NotTo(HaveOccurred())
			Expect(schema).NotTo(BeNil())
			Expect(schema.Name).To(Equal("default"))
			Expect(schema.Tables).To(HaveLen(1))

			table := schema.Tables[0]
			Expect(table.Name).To(Equal("test"))
			Expect(table.Columns).To(HaveLen(61))
			ExpectColumnsForSQLite(table.Columns)
		})

		Context("when the table has primary key", func() {
			BeforeEach(func() {
				_, err := db.Exec("CREATE TABLE my_table(id serial primary key)")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				_, err := db.Exec("DROP TABLE IF EXISTS my_table")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the schema successfully", func() {
				schema, err := provider.Schema("", "my_table")
				Expect(err).NotTo(HaveOccurred())
				Expect(schema).NotTo(BeNil())
				Expect(schema.Name).To(Equal("default"))
				Expect(schema.Tables).To(HaveLen(1))

				table := schema.Tables[0]
				Expect(table.Name).To(Equal("my_table"))
				Expect(table.Columns).To(HaveLen(1))

				column := table.Columns[0]
				Expect(column.Name).To(Equal("id"))
				Expect(column.Type.IsPrimaryKey).To(BeTrue())
			})
		})

		Context("when the table names are not provided", func() {
			It("return an error", func() {
				schema, err := provider.Schema("public")
				Expect(schema).To(BeNil())
				Expect(err).To(MatchError("No tables found"))
			})
		})

		Context("when the database is not available", func() {
			BeforeEach(func() {
				dir, err := ioutil.TempDir("", "prana")
				Expect(err).To(BeNil())

				conn := filepath.Join(dir, "prana.db")
				db, err := sqlx.Open("sqlite3", conn)
				Expect(err).NotTo(HaveOccurred())
				Expect(db.Close()).To(Succeed())

				provider.DB = db
			})

			It("return an error", func() {
				schema, err := provider.Schema("public", "test")
				Expect(schema).To(BeNil())
				Expect(err).To(MatchError("sql: database is closed"))
			})
		})
	})
})
