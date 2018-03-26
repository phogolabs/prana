package gom_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGom(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GOM Suite")
}
