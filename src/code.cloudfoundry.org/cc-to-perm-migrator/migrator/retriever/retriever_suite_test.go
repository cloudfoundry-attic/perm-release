package retriever_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRetriever(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Retriever Suite")
}
