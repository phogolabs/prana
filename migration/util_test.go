package migration_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/phogolabs/gom/fake"
	"github.com/phogolabs/gom/migration"
)

var _ = Describe("Util", func() {
	Describe("RunAll", func() {
		var (
			fs  *fake.MigrationFileSystem
			db  *sqlx.DB
			dir string
		)

		BeforeEach(func() {
			fs = &fake.MigrationFileSystem{}
			fs.WalkStub = filepath.Walk
			fs.OpenFileStub = func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
				return os.OpenFile(name, flag, perm)
			}

			var err error
			dir, err = ioutil.TempDir("", "gom_runner")
			Expect(err).To(BeNil())

			conn := filepath.Join(dir, "gom.db")
			db, err = sqlx.Open("sqlite3", conn)
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			Expect(db.Close()).To(Succeed())
		})

		It("runs all migrations successfully", func() {
			Expect(migration.RunAll(db, fs, dir)).To(Succeed())
		})

		Context("when the file system fails", func() {
			BeforeEach(func() {
				fs.WalkReturns(fmt.Errorf("Oh no!"))
				fs.OpenFileReturns(nil, fmt.Errorf("Oh no!"))
			})

			It("returns an error", func() {
				Expect(migration.RunAll(db, fs, "/migration")).To(MatchError("Oh no!"))
			})
		})
	})
})
