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

			bodyCallback func(context.Context, lager.Logger, io.Reader) error

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

			bodyCallback = func(context.Context, lager.Logger, io.Reader) error {
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

			err := c.MakePaginatedGetRequest(ctx, logger, route, bodyCallback)

			Expect(err).NotTo(HaveOccurred())

			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		It("errors when the response code is 400 or above", func() {
			route = "/some-path"

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/some-path"),
				ghttp.RespondWithJSONEncoded(400, paginatedResponse{}),
			))

			err := c.MakePaginatedGetRequest(ctx, logger, route, bodyCallback)

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

			bodyCallback = func(ctx context.Context, logger lager.Logger, r io.Reader) error {
				return errors.New("some-callback-error")
			}

			err := c.MakePaginatedGetRequest(ctx, logger, route, bodyCallback)

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
			bodyCallback = func(ctx context.Context, logger lager.Logger, r io.Reader) error {
				var p paginatedResponse

				err := json.NewDecoder(r).Decode(&p)
				Expect(err).NotTo(HaveOccurred())

				responseNames = append(responseNames, p.Name)
				return nil
			}

			err := c.MakePaginatedGetRequest(ctx, logger, route, bodyCallback)

			Expect(err).NotTo(HaveOccurred())

			Expect(server.ReceivedRequests()).To(HaveLen(3))

			Expect(responseNames).To(HaveLen(3))
			Expect(responseNames[0]).To(Equal("first"))
			Expect(responseNames[1]).To(Equal("second"))
			Expect(responseNames[2]).To(Equal("third"))
		})
	})
})

type paginatedResponse struct {
	NextURL *string `json:"next_url"`
	Name    string  `json:"name"`
}
