package integration_test

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Migration Create", func() {
	var cmd *exec.Cmd

	JustBeforeEach(func() {
		dir, err := ioutil.TempDir("", "gom")
		Expect(err).To(BeNil())

		cmd = exec.Command(gomPath, "--database-url", "sqlite3://gom.db", "migration", "create")
		cmd.Dir = dir
	})

	It("generates migration successfully", func() {
		cmd.Args = append(cmd.Args, "schema")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		path := filepath.Join(cmd.Dir, "/database/migration/*_schema.sql")
		matches, err := filepath.Glob(path)
		Expect(err).NotTo(HaveOccurred())

		Expect(matches).To(HaveLen(1))
		Expect(matches[0]).To(BeARegularFile())
	})

	Context("when the name has space in it", func() {
		It("generates migration successfully", func() {
			cmd.Args = append(cmd.Args, "my initial schema")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			path := filepath.Join(cmd.Dir, "/database/migration/*_my_initial_schema.sql")
			matches, err := filepath.Glob(path)
			Expect(err).NotTo(HaveOccurred())

			Expect(matches).To(HaveLen(1))
			Expect(matches[0]).To(BeARegularFile())
		})
	})

	Context("when the name is not provided", func() {
		It("returns an error", func() {
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(103))
			Expect(session.Err).To(gbytes.Say("Create command expects a single argument"))
		})
	})
})
