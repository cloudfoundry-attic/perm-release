package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("CcToPermMigrator", func() {
	Describe("cli", func() {
		It("exits with 1 when no flags are passed", func() {
			session := RunCommand()
			Eventually(session).Should(gexec.Exit(1))
		})
	})
})
