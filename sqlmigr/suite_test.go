package sqlmigr_test

import (
	"log"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMigration(t *testing.T) {
	log.SetOutput(GinkgoWriter)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Migration Suite")
}
