package main_test

import (
	"net"
	"os/exec"
	"strconv"

	"google.golang.org/grpc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/onsi/gomega/gexec"

	"code.cloudfoundry.org/perm-go"
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

type PermServer struct {
	hostname string
	port     int
	server   *grpc.Server
}

func NewPermServer(roleServiceServer protos.RoleServiceServer) (*PermServer, error) {
	// 0 for random port
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	addr := lis.Addr().String()
	hostname, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	iPort, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		return nil, err
	}

	server := grpc.NewServer()
	protos.RegisterRoleServiceServer(server, roleServiceServer)

	// Doesn't currently do anything special to wait for the server to actually accept connections.
	// If we notice race-related problems, we can try something like waiting for it start accepting connections.
	go server.Serve(lis)

	return &PermServer{
		hostname: hostname,
		port:     int(iPort),
		server:   server,
	}, nil
}

func (s *PermServer) Stop() {
	s.server.Stop()
}

func (s *PermServer) Hostname() string {
	return s.hostname
}

func (s *PermServer) Port() int {
	return s.port
}
