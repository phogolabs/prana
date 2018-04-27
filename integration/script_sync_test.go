package integration_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Script Sync", func() {
	var cmd *exec.Cmd

	BeforeEach(func() {
		dir, err := ioutil.TempDir("", "gom")
		Expect(err).To(BeNil())

		query := &bytes.Buffer{}
		fmt.Fprintln(query, "CREATE TABLE users (")
		fmt.Fprintln(query, "  id INT PRIMARY KEY NOT NULL,")
		fmt.Fprintln(query, "  first_name TEXT NOT NULL,")
		fmt.Fprintln(query, "  last_name TEXT")
		fmt.Fprintln(query, ");")

		db, err := sqlx.Open("sqlite3", filepath.Join(dir, "gom.db"))
		Expect(err).To(BeNil())
		_, err = db.Exec(query.String())
		Expect(err).To(BeNil())
		Expect(db.Close()).To(Succeed())

		cmd = exec.Command(gomPath, "--database-url", "sqlite3://gom.db", "migration", "setup")
		cmd.Dir = dir

		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		cmd = exec.Command(gomPath, "--database-url", "sqlite3://gom.db", "script", "sync")
		cmd.Dir = dir
	})

	It("syncs command successfully", func() {
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		path := filepath.Join(cmd.Dir, "/database/script/command.sql")
		Expect(path).To(BeARegularFile())

		data, err := ioutil.ReadFile(path)
		Expect(err).To(BeNil())

		script := string(data)
		Expect(script).To(ContainSubstring("-- name: select-all-users"))
	})
})
