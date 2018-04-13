package oak_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOAK(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OAK Suite")
}
