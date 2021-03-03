package integration_test

import (
	"os/exec"
	"testing"
	"time"

	"github.com/onsi/gomega/gexec"

	_ "github.com/mattn/go-sqlite3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var gomPath string

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	binPath, err := gexec.Build("github.com/phogolabs/prana/cmd/prana")
	Expect(err).NotTo(HaveOccurred())

	return []byte(binPath)
}, func(data []byte) {
	gomPath = string(data)
	SetDefaultEventuallyTimeout(10 * time.Second)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

func Setup(args []string, dir string) {
	cmd := exec.Command(gomPath, append(args, "migration", "setup")...)
	cmd.Dir = dir

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))

	cmd = exec.Command(gomPath, append(args, "migration", "run")...)
	cmd.Dir = dir

	session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
}
