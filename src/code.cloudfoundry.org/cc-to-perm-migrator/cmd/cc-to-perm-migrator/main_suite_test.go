package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/onsi/gomega/gexec"
)

func TestCcToPermMigrator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CcToPermMigrator Suite")
}

var cliPath string

var _ = SynchronizedBeforeSuite(func() []byte {
	path, err := gexec.Build("code.cloudfoundry.org/cc-to-perm-migrator/cmd/cc-to-perm-migrator")
	Expect(err).NotTo(HaveOccurred())

	return []byte(path)
}, func(data []byte) {
	cliPath = string(data)
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
})

func RunCommand(args ...string) *gexec.Session {
	cmd := exec.Command(cliPath, args...)

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	<-session.Exited

	return session
}
