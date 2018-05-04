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

var _ = Describe("Migration Run", func() {
	var (
		cmd *exec.Cmd
		db  *sql.DB
	)

	JustBeforeEach(func() {
		dir, err := ioutil.TempDir("", "gom")
		Expect(err).To(BeNil())

		args := []string{"--database-url", "sqlite3://gom.db"}

		Setup(args, dir)

		args = append(args, "migration")

		script := &bytes.Buffer{}
		fmt.Fprintln(script, "-- name: up")
		fmt.Fprintln(script, "SELECT * FROM migrations;")
		fmt.Fprintln(script, "-- name: down")
		fmt.Fprintln(script, "SELECT * FROM migrations;")

		path := filepath.Join(dir, "/database/migration/20060102150405_schema.sql")
		Expect(ioutil.WriteFile(path, script.Bytes(), 0700)).To(Succeed())

		path = filepath.Join(dir, "/database/migration/20070102150405_trigger.sql")
		Expect(ioutil.WriteFile(path, script.Bytes(), 0700)).To(Succeed())

		cmd = exec.Command(gomPath, append(args, "run")...)
		cmd.Dir = dir

		db, err = sql.Open("sqlite3", filepath.Join(dir, "gom.db"))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(db.Close()).To(Succeed())
	})

	It("runs migration successfully", func() {
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		row := db.QueryRow("SELECT COUNT(*) FROM migrations")

		count := 0
		Expect(row.Scan(&count)).To(Succeed())
		Expect(count).To(Equal(2))
	})

	Context("when the count argument is provided", func() {
		It("runs migration successfully", func() {
			cmd.Args = append(cmd.Args, "--count", "2")

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			row := db.QueryRow("SELECT COUNT(*) FROM migrations")

			count := 0
			Expect(row.Scan(&count)).To(Succeed())
			Expect(count).To(Equal(3))
		})

		Context("when the count is negative number", func() {
			It("returns an error", func() {
				cmd.Args = append(cmd.Args, "--count", "-1")

				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(103))
				Expect(session.Err).To(gbytes.Say("The count argument cannot be negative number"))
			})
		})
	})

	Context("when the database is not available", func() {
		It("returns an error", func() {
			Expect(os.Remove(filepath.Join(cmd.Dir, "gom.db"))).To(Succeed())

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
		})
	})

	Context("when the count argument is wrong", func() {
		It("returns an error", func() {
			cmd.Args = append(cmd.Args, "--count", "wrong")

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(-1))
			Expect(session.Out).To(gbytes.Say(`Incorrect Usage: invalid value "wrong" for flag -count: strconv.ParseInt: parsing "wrong": invalid syntax`))
		})
	})
})
