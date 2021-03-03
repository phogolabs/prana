package integration_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Script Create", func() {
	var cmd *exec.Cmd

	JustBeforeEach(func() {
		dir, err := ioutil.TempDir("", "gom")
		Expect(err).To(BeNil())

		cmd = exec.Command(gomPath, "routine", "create")
		cmd.Dir = dir

		Expect(os.MkdirAll(filepath.Join(dir, "database", "routine"), 0700)).To(Succeed())
	})

	It("generates command successfully", func() {
		cmd.Args = append(cmd.Args, "-n", "commands", "update-user")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		path := filepath.Join(cmd.Dir, "/database/routine/commands.sql")
		Expect(path).To(BeARegularFile())

		data, err := ioutil.ReadFile(path)
		Expect(err).To(BeNil())

		script := string(data)
		Expect(script).To(ContainSubstring("-- name: update-user"))
	})

	Context("when the command name has space in it", func() {
		It("generates command successfully", func() {
			cmd.Args = append(cmd.Args, "-n", "commands", "update user")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			path := filepath.Join(cmd.Dir, "/database/routine/commands.sql")
			Expect(path).To(BeARegularFile())

			data, err := ioutil.ReadFile(path)
			Expect(err).To(BeNil())

			script := string(data)
			Expect(script).To(ContainSubstring("-- name: update-user"))
		})
	})

	Context("when the container name has space in it", func() {
		It("generates command successfully", func() {
			cmd.Args = append(cmd.Args, "-n", "my commands", "update-user")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			path := filepath.Join(cmd.Dir, "/database/routine/my_commands.sql")
			Expect(path).To(BeARegularFile())

			data, err := ioutil.ReadFile(path)
			Expect(err).To(BeNil())

			script := string(data)
			Expect(script).To(ContainSubstring("-- name: update-user"))
		})
	})

	Context("when the name is not provided", func() {
		It("returns an error", func() {
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(104))
			Expect(session.Err).To(gbytes.Say("Create command expects a single argument"))
		})
	})
})
