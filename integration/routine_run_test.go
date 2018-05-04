package integration_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Script Run", func() {
	var (
		cmd *exec.Cmd
		db  *sql.DB
	)

	JustBeforeEach(func() {
		dir, err := ioutil.TempDir("", "gom")
		Expect(err).To(BeNil())

		args := []string{"--database-url", "sqlite3://gom.db"}

		Setup(args, dir)

		script := &bytes.Buffer{}
		fmt.Fprintln(script, "-- name: show-migrations")
		fmt.Fprintln(script, "SELECT * FROM migrations;")

		Expect(os.MkdirAll(filepath.Join(dir, "/database/routine"), 0700)).To(Succeed())
		path := filepath.Join(dir, "/database/routine/20060102150405.sql")
		Expect(ioutil.WriteFile(path, script.Bytes(), 0700)).To(Succeed())

		cmd = exec.Command(gomPath, append(args, "routine", "run")...)
		cmd.Dir = dir

		db, err = sql.Open("sqlite3", filepath.Join(dir, "gom.db"))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(db.Close()).To(Succeed())
	})

	It("runs command successfully", func() {
		cmd.Args = append(cmd.Args, "show-migrations")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		Expect(session.Err).To(gbytes.Say("Running command 'show-migrations'"))
	})

	Context("when the database is not available", func() {
		It("returns an error", func() {
			Expect(os.Remove(filepath.Join(cmd.Dir, "gom.db"))).To(Succeed())

			cmd.Args = append(cmd.Args, "show-migrations")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(-1))
			Expect(session.Err).To(gbytes.Say("no such table: migrations"))
		})
	})

	Context("when the command name is missing", func() {
		It("returns an error", func() {
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(-1))
			Expect(session.Err).To(gbytes.Say("Run command expects a single argument"))
		})
	})
})
