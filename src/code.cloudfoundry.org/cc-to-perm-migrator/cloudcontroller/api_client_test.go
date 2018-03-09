package cloudcontroller_test

import (
	"errors"

	. "code.cloudfoundry.org/cc-to-perm-migrator/cloudcontroller"

	"context"
	"io"

	"net/http"

	"encoding/json"

	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("APIClient", func() {
	Describe("#c.MakePaginatedGetRequest", func() {
		var (
			c *APIClient

			ctx     context.Context
			client  *http.Client
			timeout time.Duration

			logger *lagertest.TestLogger

			host  string
			route string

			bodyCallback func(lager.Logger, io.Reader) error

			server *ghttp.Server
		)

		BeforeEach(func() {
			server = ghttp.NewServer()

			ctx = context.Background()

			logger = lagertest.NewTestLogger("make-paginated-get-request")

			client = http.DefaultClient

			timeout = 5 * time.Second

			host = server.URL()
			route = ""

			bodyCallback = func(lager.Logger, io.Reader) error {
				return nil
			}

			c = NewAPIClient(host, client, timeout)
		})

		AfterEach(func() {
			server.Close()
		})

		It("makes a JSON API get request", func() {
			route = "/some-path"

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/some-path"),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.RespondWithJSONEncoded(200, paginatedResponse{}),
			))

			err := c.MakePaginatedGetRequest(logger, route, bodyCallback)

			Expect(err).NotTo(HaveOccurred())

			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		It("errors when the response code is 400 or above", func() {
			route = "/some-path"

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/some-path"),
				ghttp.RespondWithJSONEncoded(400, paginatedResponse{}),
			))

			err := c.MakePaginatedGetRequest(logger, route, bodyCallback)

			Expect(err).To(HaveOccurred())
		})

		It("errors when there's an error in the callback", func() {
			route = "/some-path"

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/some-path"),
				ghttp.RespondWithJSONEncoded(200, paginatedResponse{
					Name: "first",
				}),
			))

			bodyCallback = func(logger lager.Logger, r io.Reader) error {
				return errors.New("some-callback-error")
			}

			err := c.MakePaginatedGetRequest(logger, route, bodyCallback)

			Expect(err).To(MatchError("some-callback-error"))
		})

		It("makes multiple requests until the responses no longer have a next URL, calling bodyCallback for each with a copy of the body", func() {
			route = "/some-path"

			nextURL1 := "/some-other-path"
			nextURL2 := "/some-third-path"
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/some-path"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.RespondWithJSONEncoded(200, paginatedResponse{
						NextURL: &nextURL1,
						Name:    "first",
					}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/some-other-path"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.RespondWithJSONEncoded(200, paginatedResponse{
						NextURL: &nextURL2,
						Name:    "second",
					}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/some-third-path"),
					ghttp.VerifyHeaderKV("Accept", "application/json"),
					ghttp.RespondWithJSONEncoded(200, paginatedResponse{
						NextURL: nil,
						Name:    "third",
					}),
				),
			)

			var responseNames []string
			bodyCallback = func(logger lager.Logger, r io.Reader) error {
				var p paginatedResponse

				err := json.NewDecoder(r).Decode(&p)
				Expect(err).NotTo(HaveOccurred())

				responseNames = append(responseNames, p.Name)
				return nil
			}

			err := c.MakePaginatedGetRequest(logger, route, bodyCallback)

			Expect(err).NotTo(HaveOccurred())

			Expect(server.ReceivedRequests()).To(HaveLen(3))

			Expect(responseNames).To(HaveLen(3))
			Expect(responseNames[0]).To(Equal("first"))
			Expect(responseNames[1]).To(Equal("second"))
			Expect(responseNames[2]).To(Equal("third"))
		})
	})

	Describe("#c.GetOrganizations", func() {
		var (
			c *APIClient

			ctx     context.Context
			client  *http.Client
			timeout time.Duration

			logger *lagertest.TestLogger

			host   string
			route  string
			server *ghttp.Server
		)

		BeforeEach(func() {
			server = ghttp.NewServer()
			ctx = context.Background()
			logger = lagertest.NewTestLogger("get-organizations")
			client = http.DefaultClient

			timeout = 5 * time.Second
			host = server.URL()
			route = ""
			c = NewAPIClient(host, client, timeout)
		})

		It("returns a list of organizations", func() {
			route = "/v2/organizations"

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", route),
				ghttp.VerifyHeaderKV("Accept", "application/json"),
				ghttp.RespondWith(200, `{
			  "total_results": 2,
			  "total_pages": 1,
			  "prev_url": null,
			  "next_url": null,
			  "resources": [
				{
				  "metadata": {
					"guid": "org1-guid",
					"url": "/v2/organizations/org1-guid",
					"created_at": "2016-06-08T16:41:33Z",
					"updated_at": "2016-06-08T16:41:26Z"
				  },
				  "entity": {
					"name": "org1",
					"billing_enabled": false,
					"quota_definition_guid": "dcb680a9-b190-4838-a3d2-b84aa17517a6",
					"status": "active",
					"quota_definition_url": "/v2/quota_definitions/dcb680a9-b190-4838-a3d2-b84aa17517a6",
					"spaces_url": "/v2/organizations/org1-guid/spaces",
					"domains_url": "/v2/organizations/org1-guid/domains",
					"private_domains_url": "/v2/organizations/org1-guid/private_domains",
					"users_url": "/v2/organizations/org1-guid/users",
					"managers_url": "/v2/organizations/org1-guid/managers",
					"billing_managers_url": "/v2/organizations/org1-guid/billing_managers",
					"auditors_url": "/v2/organizations/org1-guid/auditors",
					"app_events_url": "/v2/organizations/org1-guid/app_events",
					"space_quota_definitions_url": "/v2/organizations/org1-guid/space_quota_definitions"
				  }
				},
				{
				  "metadata": {
					"guid": "org2-guid",
					"url": "/v2/organizations/org2-guid",
					"created_at": "2016-06-08T16:41:33Z",
					"updated_at": "2016-06-08T16:41:26Z"
				  },
				  "entity": {
					"name": "org2",
					"billing_enabled": false,
					"quota_definition_guid": "dcb680a9-b190-4838-a3d2-b84aa17517a6",
					"status": "active",
					"quota_definition_url": "/v2/quota_definitions/dcb680a9-b190-4838-a3d2-b84aa17517a6",
					"spaces_url": "/v2/organizations/org2-guid/spaces",
					"domains_url": "/v2/organizations/org2-guid/domains",
					"private_domains_url": "/v2/organizations/org2-guid/private_domains",
					"users_url": "/v2/organizations/org2-guid/users",
					"managers_url": "/v2/organizations/org2-guid/managers",
					"billing_managers_url": "/v2/organizations/org2-guid/billing_managers",
					"auditors_url": "/v2/organizations/org2-guid/auditors",
					"app_events_url": "/v2/organizations/org2-guid/app_events",
					"space_quota_definitions_url": "/v2/organizations/org2-guid/space_quota_definitions"
				  }
				}
			  ]
			}`)))

			orgs, err := c.GetOrganizations(logger)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(*orgs)).To(Equal(2))
			for i, expectedOrg := range []string{"org1-guid", "org2-guid"} {
				Expect(expectedOrg).To(Equal((*orgs)[i].Metadata.GUID))
			}
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})
	})
})

type paginatedResponse struct {
	NextURL *string `json:"next_url"`
	Name    string  `json:"name"`
}
