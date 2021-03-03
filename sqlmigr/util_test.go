package sqlmigr_test

import (
	"io/ioutil"
	"path/filepath"
	"testing/fstest"

	"github.com/jmoiron/sqlx"
	"github.com/phogolabs/prana/sqlmigr"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Util", func() {
	Describe("RunAll", func() {
		var db *sqlx.DB

		BeforeEach(func() {
			dir, err := ioutil.TempDir("", "prana_runner")
			Expect(err).To(BeNil())

			conn := filepath.Join(dir, "prana.db")
			db, err = sqlx.Open("sqlite3", conn)
			Expect(err).To(BeNil())

		})

		AfterEach(func() {
			Expect(db.Close()).To(Succeed())
		})

		It("runs all sqlmigrs successfully", func() {
			Expect(sqlmigr.RunAll(db, fstest.MapFS{})).To(Succeed())
		})
	})
})
