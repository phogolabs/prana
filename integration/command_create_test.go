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

var _ = Describe("Command Create", func() {
	var cmd *exec.Cmd

	JustBeforeEach(func() {
		dir, err := ioutil.TempDir("", "gom")
		Expect(err).To(BeNil())

		cmd = exec.Command(gomPath, "command", "create")
		cmd.Dir = dir
	})

	It("generates command successfully", func() {
		cmd.Args = append(cmd.Args, "commands", "update-user")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		path := filepath.Join(cmd.Dir, "/database/command/commands.sql")
		Expect(path).To(BeARegularFile())

		data, err := ioutil.ReadFile(path)
		Expect(err).To(BeNil())

		script := string(data)
		Expect(script).To(ContainSubstring("-- name: update-user"))
	})

	Context("when the command name has space in it", func() {
		It("generates command successfully", func() {
			cmd.Args = append(cmd.Args, "commands", "update user")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			path := filepath.Join(cmd.Dir, "/database/command/commands.sql")
			Expect(path).To(BeARegularFile())

			data, err := ioutil.ReadFile(path)
			Expect(err).To(BeNil())

			script := string(data)
			Expect(script).To(ContainSubstring("-- name: update-user"))
		})
	})

	Context("when the container name has space in it", func() {
		It("generates command successfully", func() {
			cmd.Args = append(cmd.Args, "my commands", "update-user")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			path := filepath.Join(cmd.Dir, "/database/command/my_commands.sql")
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
			Expect(session.Err).To(gbytes.Say("Create command expects two argument"))
		})
	})
})
