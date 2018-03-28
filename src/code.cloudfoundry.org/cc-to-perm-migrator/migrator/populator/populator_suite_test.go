package populator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPopulator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Populator Suite")
}
