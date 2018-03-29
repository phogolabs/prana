package integration_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Migration Status", func() {
	var (
		cmd *exec.Cmd
		db  *sql.DB
	)

	JustBeforeEach(func() {
		dir, err := ioutil.TempDir("", "gom")
		Expect(err).To(BeNil())

		args := []string{"--database-url", "sqlite3://gom.db", "migration"}

		cmd = exec.Command(gomPath, append(args, "setup")...)
		cmd.Dir = dir

		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		script := &bytes.Buffer{}
		fmt.Fprintln(script, "-- name: up")
		fmt.Fprintln(script, "SELECT * FROM migrations;")
		fmt.Fprintln(script, "-- name: down")
		fmt.Fprintln(script, "SELECT * FROM migrations;")

		path := filepath.Join(cmd.Dir, "/database/migration/20060102150405_schema.sql")
		Expect(ioutil.WriteFile(path, script.Bytes(), 0700)).To(Succeed())

		cmd = exec.Command(gomPath, append(args, "status")...)
		cmd.Dir = dir

		db, err = sql.Open("sqlite3", filepath.Join(dir, "gom.db"))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(db.Close()).To(Succeed())
	})

	It("returns the migration status successfully", func() {
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		row := db.QueryRow("SELECT created_at FROM migrations WHERE id = '00060524000000'")

		timestamp := time.Now()
		Expect(row.Scan(&timestamp)).To(Succeed())

		Expect(string(session.Out.Contents())).To(ContainSubstring("00060524000000"))
		Expect(string(session.Out.Contents())).To(ContainSubstring("20060102150405"))
	})

	Context("when the database is not available", func() {
		It("returns an error", func() {
			Expect(os.Remove(filepath.Join(cmd.Dir, "gom.db"))).To(Succeed())

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(-1))
		})
	})
})
