package sqlmigr_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/phogolabs/parcello"
	"github.com/phogolabs/prana/sqlmigr"
)

var _ = Describe("Util", func() {
	Describe("RunAll", func() {
		var (
			db *sqlx.DB
			fs sqlmigr.FileSystem
		)

		BeforeEach(func() {
			dir, err := ioutil.TempDir("", "prana_runner")
			Expect(err).To(BeNil())

			conn := filepath.Join(dir, "prana.db")
			db, err = sqlx.Open("sqlite3", conn)
			Expect(err).To(BeNil())

			fs = parcello.Dir(dir)
		})

		AfterEach(func() {
			Expect(db.Close()).To(Succeed())
		})

		It("runs all sqlmigrs successfully", func() {
			Expect(sqlmigr.RunAll(db, fs)).To(Succeed())
		})

		Context("when the file system fails", func() {
			BeforeEach(func() {
				fs = parcello.Dir("/file")
			})

			It("returns an error", func() {
				Expect(sqlmigr.RunAll(db, fs)).To(MatchError(os.ErrNotExist))
			})
		})
	})
})
