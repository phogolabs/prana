package sqlexec_test

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestScript(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SQLExec Suite")
}
