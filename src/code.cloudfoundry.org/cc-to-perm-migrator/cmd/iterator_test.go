package cmd_test

import (
	. "code.cloudfoundry.org/cc-to-perm-migrator/cmd"

	"context"
	"io"

	"strings"

	"fmt"

	"code.cloudfoundry.org/cc-to-perm-migrator/cmd/cmdfakes"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe(".IterateOverCloudControllerEntities", func() {
	var (
		ctx context.Context

		logger *lagertest.TestLogger

		c chan RoleAssignment

		ccAPIClient *cmdfakes.FakeCloudControllerAPIClient

		routeResponses map[string]string
	)

	BeforeEach(func() {
		ctx = context.Background()

		logger = lagertest.NewTestLogger("iterate-over-cloud-controller-entities")

		c = make(chan RoleAssignment, 1000)

		routeResponses = make(map[string]string)

		ccAPIClient = new(cmdfakes.FakeCloudControllerAPIClient)

		ccAPIClient.MakePaginatedGetRequestStub = func(ctx context.Context, logger lager.Logger, route string, bodyCallback func(context.Context, lager.Logger, io.Reader) error) error {
			response, ok := routeResponses[route]
			if !ok {
				return fmt.Errorf("Expected to find response for route %s", route)
			}

			return bodyCallback(ctx, logger, strings.NewReader(response))
		}
	})

	Context("when no organizations", func() {
		JustBeforeEach(func() {
			routeResponses["/v2/organizations"] = `{}`
		})

		It("hits /v2/organizations", func() {
			err := IterateOverCloudControllerEntities(ctx, logger, c, ccAPIClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(ccAPIClient.MakePaginatedGetRequestCallCount()).To(Equal(1))

			var route string
			_, _, route, _ = ccAPIClient.MakePaginatedGetRequestArgsForCall(0)
			Expect(route).To(Equal("/v2/organizations"))
		})
	})

	Context("when there are organizations", func() {
		JustBeforeEach(func() {
			routeResponses["/v2/organizations"] = `{
			  "total_results": 1,
			  "total_pages": 1,
			  "prev_url": null,
			  "next_url": null,
			  "resources": [
				{
				  "metadata": {
					"guid": "a7aff246-5f5b-4cf8-87d8-f316053e4a20",
					"url": "/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20",
					"created_at": "2016-06-08T16:41:33Z",
					"updated_at": "2016-06-08T16:41:26Z"
				  },
				  "entity": {
					"name": "the-system_domain-org-name",
					"billing_enabled": false,
					"quota_definition_guid": "dcb680a9-b190-4838-a3d2-b84aa17517a6",
					"status": "active",
					"quota_definition_url": "/v2/quota_definitions/dcb680a9-b190-4838-a3d2-b84aa17517a6",
					"spaces_url": "/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/spaces",
					"domains_url": "/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/domains",
					"private_domains_url": "/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/private_domains",
					"users_url": "/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/users",
					"managers_url": "/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/managers",
					"billing_managers_url": "/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/billing_managers",
					"auditors_url": "/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/auditors",
					"app_events_url": "/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/app_events",
					"space_quota_definitions_url": "/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/space_quota_definitions"
				  }
				}
			  ]
			}`

			routeResponses["/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/spaces"] = `{
			  "total_results": 1,
			  "total_pages": 1,
			  "prev_url": null,
			  "next_url": null,
			  "resources": [
				{
				  "metadata": {
					"guid": "5489e195-c42b-4e61-bf30-323c331ecc01",
					"url": "/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01",
					"created_at": "2016-06-08T16:41:35Z",
					"updated_at": "2016-06-08T16:41:26Z"
				  },
				  "entity": {
					"name": "name-1774",
					"organization_guid": "3deb9f04-b449-4f94-b3dd-c73cefe5b275",
					"space_quota_definition_guid": null,
					"allow_ssh": true,
					"organization_url": "/v2/organizations/3deb9f04-b449-4f94-b3dd-c73cefe5b275",
					"developers_url": "/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/developers",
					"managers_url": "/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/managers",
					"auditors_url": "/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/auditors",
					"apps_url": "/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/apps",
					"routes_url": "/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/routes",
					"domains_url": "/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/domains",
					"service_instances_url": "/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/service_instances",
					"app_events_url": "/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/app_events",
					"events_url": "/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/events",
					"security_groups_url": "/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/security_groups",
					"staging_security_groups_url": "/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/staging_security_groups"
				  }
				}
			  ]
			}`

			routeResponses["/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/users"] = `{}`
			routeResponses["/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/billing_managers"] = `{}`
			routeResponses["/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/managers"] = `{}`
			routeResponses["/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/auditors"] = `{}`

			routeResponses["/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/developers"] = `{}`
			routeResponses["/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/auditors"] = `{}`
			routeResponses["/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/managers"] = `{}`

		})

		It("hits the spaces, users, billing managers, managers, and auditors URLs for every organization, and the developers, auditors, and managers URL for every space", func() {
			err := IterateOverCloudControllerEntities(ctx, logger, c, ccAPIClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(ccAPIClient.MakePaginatedGetRequestCallCount()).To(Equal(9))

			var route string
			_, _, route, _ = ccAPIClient.MakePaginatedGetRequestArgsForCall(0)
			Expect(route).To(Equal("/v2/organizations"))

			_, _, route, _ = ccAPIClient.MakePaginatedGetRequestArgsForCall(1)
			Expect(route).To(Equal("/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/spaces"))

			requestCount := ccAPIClient.MakePaginatedGetRequestCallCount()

			var routes []string
			for i := 2; i < requestCount; i++ {
				_, _, route, _ = ccAPIClient.MakePaginatedGetRequestArgsForCall(i)
				routes = append(routes, route)
			}

			Expect(routes).To(ContainElement("/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/users"))
			Expect(routes).To(ContainElement("/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/billing_managers"))
			Expect(routes).To(ContainElement("/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/managers"))
			Expect(routes).To(ContainElement("/v2/organizations/a7aff246-5f5b-4cf8-87d8-f316053e4a20/auditors"))
			Expect(routes).To(ContainElement("/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/developers"))
			Expect(routes).To(ContainElement("/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/auditors"))
			Expect(routes).To(ContainElement("/v2/spaces/5489e195-c42b-4e61-bf30-323c331ecc01/managers"))
		})
	})
})
