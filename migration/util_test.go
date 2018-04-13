package migration_test

import (
	"io/ioutil"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/phogolabs/gom/migration"
	"github.com/phogolabs/parcel"
)

var _ = Describe("Util", func() {
	Describe("RunAll", func() {
		var (
			db *sqlx.DB
			fs migration.FileSystem
		)

		BeforeEach(func() {
			dir, err := ioutil.TempDir("", "gom_runner")
			Expect(err).To(BeNil())

			conn := filepath.Join(dir, "gom.db")
			db, err = sqlx.Open("sqlite3", conn)
			Expect(err).To(BeNil())

			fs = parcel.Dir(dir)
		})

		AfterEach(func() {
			Expect(db.Close()).To(Succeed())
		})

		It("runs all migrations successfully", func() {
			Expect(migration.RunAll(db, fs)).To(Succeed())
		})

		Context("when the file system fails", func() {
			BeforeEach(func() {
				fs = parcel.Dir("/")
			})

			It("returns an error", func() {
				Expect(migration.RunAll(db, fs)).To(MatchError("open /00060524000000_setup.sql: permission denied"))
			})
		})
	})
})
