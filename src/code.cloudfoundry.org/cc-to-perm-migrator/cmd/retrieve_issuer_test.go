package cmd_test

import (
	"context"

	. "code.cloudfoundry.org/cc-to-perm-migrator/cmd"

	"net/http"

	"net/url"

	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("RetrieveIssuer", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})

	It("fetches the issuer from the JSON of .well-known/openid-configuration", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/.well-known/openid-configuration"),
				ghttp.RespondWith(200, `{"issuer": "foo"}`),
			),
		)

		u, err := url.Parse(server.URL())
		Expect(err).NotTo(HaveOccurred())
		issuer, err := RetrieveIssuer(context.Background(), lagertest.NewTestLogger("retrieve-issuer"), http.DefaultClient, u)

		Expect(err).NotTo(HaveOccurred())
		Expect(issuer).To(Equal("foo"))
	})
})
