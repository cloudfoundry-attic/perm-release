package capi_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"net/http"

	. "code.cloudfoundry.org/cc-to-perm-migrator/capi"
	"code.cloudfoundry.org/cc-to-perm-migrator/capi/capimodels"
	"code.cloudfoundry.org/cc-to-perm-migrator/migrator"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Client", func() {
	var (
		logger *lagertest.TestLogger
		server *ghttp.Server
		client *Client
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test-capi-client")
		server = ghttp.NewServer()
		client = NewClient(server.URL())
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("#GetOrgRoleAssignments", func() {
		var (
			orgGUID string

			page1Path string
			page1     capimodels.ListOrgRolesResponse
			page2Path string
			page2     capimodels.ListOrgRolesResponse
		)

		BeforeEach(func() {
			orgGUID = "test-org-guid"
			page1Path = fmt.Sprintf("/v2/organizations/%s/user_roles", orgGUID)
			page2Path = fmt.Sprintf("/FAKE-NEXT-PAGE-PATH")

			page1 = capimodels.ListOrgRolesResponse{
				PaginatedResponse: capimodels.PaginatedResponse{
					NextURL: &page2Path,
				},
				Resources: []capimodels.OrgUserResource{
					{
						Metadata: capimodels.MetadataResource{
							GUID: "test-user-1",
						},
						Entity: capimodels.OrgUserResourceEntity{
							Roles: []string{"org_developer"},
						},
					},
					{
						Metadata: capimodels.MetadataResource{
							GUID: "test-user-2",
						},
						Entity: capimodels.OrgUserResourceEntity{
							Roles: []string{"org_auditor", "billing_manager"},
						},
					},
				},
			}

			page2 = capimodels.ListOrgRolesResponse{
				PaginatedResponse: capimodels.PaginatedResponse{
					NextURL: nil,
				},
				Resources: []capimodels.OrgUserResource{
					{
						Metadata: capimodels.MetadataResource{
							GUID: "test-user-3",
						},
						Entity: capimodels.OrgUserResourceEntity{
							Roles: []string{"org_manager"},
						},
					},
				},
			}
		})

		Context("when the server responds successfully", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", page1Path),
						ghttp.RespondWithJSONEncoded(http.StatusOK, page1),
					),
				)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", page2Path),
						ghttp.RespondWithJSONEncoded(http.StatusOK, page2),
					),
				)
			})

			It("should return a list of org assignments", func() {
				roleAssignments, err := client.GetOrgRoleAssignments(logger, orgGUID)

				Expect(server.ReceivedRequests()).To(HaveLen(2))
				Expect(err).NotTo(HaveOccurred())

				expectedResources := append(page1.Resources, page2.Resources...)

				Expect(roleAssignments).To(HaveLen(len(expectedResources)))

				for _, resource := range expectedResources {
					Expect(roleAssignments).To(ContainElement(migrator.RoleAssignment{
						UserGUID:     resource.Metadata.GUID,
						ResourceGUID: orgGUID,
						Roles:        resource.Entity.Roles,
					}))
				}
			})
		})

		Context("when the server returns an error on the first page", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", page1Path),
						ghttp.RespondWithJSONEncoded(http.StatusInternalServerError, ""),
					),
				)
			})

			It("should return an empty list of orgs and the error", func() {
				roleAssignments, err := client.GetOrgRoleAssignments(logger, orgGUID)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(roleAssignments).To(BeEmpty())

				Expect(err).To(MatchError("failed-to-fetch-organization-user-roles"))
			})
		})

		Context("when the server returns an error on the nth page", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", page1Path),
						ghttp.RespondWithJSONEncoded(http.StatusOK, page1),
					),
				)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", page2Path),
						ghttp.RespondWithJSONEncoded(http.StatusInternalServerError, "oh no"),
					),
				)
			})

			It("should return the orgs from the n-1 pages and the error", func() {
				roleAssignments, err := client.GetOrgRoleAssignments(logger, orgGUID)

				Expect(server.ReceivedRequests()).To(HaveLen(2))

				Expect(err).To(MatchError("failed-to-fetch-organization-user-roles"))

				expectedResources := page1.Resources
				Expect(roleAssignments).To(HaveLen(len(expectedResources)))

				for _, resource := range expectedResources {
					Expect(roleAssignments).To(ContainElement(migrator.RoleAssignment{
						UserGUID:     resource.Metadata.GUID,
						ResourceGUID: orgGUID,
						Roles:        resource.Entity.Roles,
					}))
				}

			})
		})

		Context("when the response contains bad JSON", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", page1Path),
						ghttp.VerifyHeaderKV("Accept", "application/json"),
						ghttp.RespondWith(http.StatusOK, "bad response"),
					),
				)
			})

			It("should return an empty list of orgs and an error", func() {
				roleAssignments, err := client.GetOrgRoleAssignments(logger, orgGUID)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(roleAssignments).To(BeEmpty())

				Expect(err).To(MatchError("failed-to-fetch-organization-user-roles"))
			})
		})
	})

	Describe("#GetSpaceRoleAssignments", func() {
		var (
			spaceGUID string

			page1Path string
			page1     capimodels.ListSpaceRolesResponse
			page2Path string
			page2     capimodels.ListSpaceRolesResponse
		)

		BeforeEach(func() {
			spaceGUID = "test-space-guid"

			page1Path = fmt.Sprintf("/v2/spaces/%s/user_roles", spaceGUID)
			page2Path = fmt.Sprintf("/FAKE-NEXT-PAGE-PATH")

			page1 = capimodels.ListSpaceRolesResponse{
				PaginatedResponse: capimodels.PaginatedResponse{
					NextURL: &page2Path,
				},
				Resources: []capimodels.SpaceUserResource{
					{
						Metadata: capimodels.MetadataResource{
							GUID: "test-user-1",
						},
						Entity: capimodels.SpaceUserResourceEntity{
							Roles: []string{"space_developer"},
						},
					},
					{
						Metadata: capimodels.MetadataResource{
							GUID: "test-user-2",
						},
						Entity: capimodels.SpaceUserResourceEntity{
							Roles: []string{"space_auditor", "space_manager"},
						},
					},
				},
			}

			page2 = capimodels.ListSpaceRolesResponse{
				PaginatedResponse: capimodels.PaginatedResponse{
					NextURL: nil,
				},
				Resources: []capimodels.SpaceUserResource{
					{
						Metadata: capimodels.MetadataResource{
							GUID: "test-user-3",
						},
						Entity: capimodels.SpaceUserResourceEntity{
							Roles: []string{"space-manager"},
						},
					},
				},
			}
		})

		Context("when the server responds successfully", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", page1Path),
						ghttp.RespondWithJSONEncoded(http.StatusOK, page1),
					),
				)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", page2Path),
						ghttp.RespondWithJSONEncoded(http.StatusOK, page2),
					),
				)
			})

			It("should return a list of space assignments", func() {
				roleAssignments, err := client.GetSpaceRoleAssignments(logger, spaceGUID)

				Expect(server.ReceivedRequests()).To(HaveLen(2))
				Expect(err).NotTo(HaveOccurred())

				expectedResources := append(page1.Resources, page2.Resources...)

				Expect(roleAssignments).To(HaveLen(len(expectedResources)))

				for _, resource := range expectedResources {
					Expect(roleAssignments).To(ContainElement(migrator.RoleAssignment{
						UserGUID:     resource.Metadata.GUID,
						ResourceGUID: spaceGUID,
						Roles:        resource.Entity.Roles,
					}))
				}
			})
		})

		Context("when the server returns an error on the first page", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", page1Path),
						ghttp.RespondWithJSONEncoded(http.StatusInternalServerError, ""),
					),
				)
			})

			It("should return an empty list of spaces and the error", func() {
				roleAssignments, err := client.GetSpaceRoleAssignments(logger, spaceGUID)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(roleAssignments).To(BeEmpty())

				Expect(err).To(MatchError("failed-to-fetch-space-user-roles"))
			})
		})

		Context("when the server returns an error on the nth page", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", page1Path),
						ghttp.RespondWithJSONEncoded(http.StatusOK, page1),
					),
				)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", page2Path),
						ghttp.RespondWithJSONEncoded(http.StatusInternalServerError, "oh no"),
					),
				)
			})

			It("should return the orgs from the n-1 pages and the error", func() {
				roleAssignments, err := client.GetSpaceRoleAssignments(logger, spaceGUID)

				Expect(server.ReceivedRequests()).To(HaveLen(2))

				Expect(err).To(MatchError("failed-to-fetch-space-user-roles"))

				expectedResources := page1.Resources
				Expect(roleAssignments).To(HaveLen(len(expectedResources)))

				for _, resource := range expectedResources {
					Expect(roleAssignments).To(ContainElement(migrator.RoleAssignment{
						UserGUID:     resource.Metadata.GUID,
						ResourceGUID: spaceGUID,
						Roles:        resource.Entity.Roles,
					}))
				}

			})
		})

		Context("when the response contains bad JSON", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", page1Path),
						ghttp.VerifyHeaderKV("Accept", "application/json"),
						ghttp.RespondWith(http.StatusOK, "bad response"),
					),
				)
			})

			It("should return an empty list of orgs and an error", func() {
				roleAssignments, err := client.GetSpaceRoleAssignments(logger, spaceGUID)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(roleAssignments).To(BeEmpty())

				Expect(err).To(MatchError("failed-to-fetch-space-user-roles"))
			})
		})
	})
})
