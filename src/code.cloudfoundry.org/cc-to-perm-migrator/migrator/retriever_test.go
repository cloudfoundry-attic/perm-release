package migrator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"

	. "code.cloudfoundry.org/cc-to-perm-migrator/migrator"
	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/migratorfakes"
	"code.cloudfoundry.org/lager/lagertest"
)

var _ = Describe("Retriever", func() {
	var assignments chan RoleAssignment
	var errs chan error
	var client *migratorfakes.FakeCAPIClient
	var logger *lagertest.TestLogger

	BeforeEach(func() {
		assignments = make(chan RoleAssignment, 10)
		errs = make(chan error, 10)
		client = new(migratorfakes.FakeCAPIClient)
		logger = lagertest.NewTestLogger("fetch-capi-entities")
	})

	Describe("#FetchCAPIEntities", func() {
		BeforeEach(func() {
			client.GetOrgGUIDsReturns([]string{"org-guid"}, nil)
		})
		Context("when the capi client returns a single org with an org role assignment", func() {
			BeforeEach(func() {
				client.GetOrgRoleAssignmentsReturns(
					[]RoleAssignment{
						{
							ResourceGUID: "org-guid",
							UserGUID:     "user-guid",
							Roles:        []string{"org_auditor"},
						},
					}, nil)

			})
			It("returns an org auditor assignment to the channel", func() {
				FetchCAPIEntities(client, logger, assignments, errs)

				Expect(client.GetOrgGUIDsCallCount()).To(Equal(1))
				_, orgGUID := client.GetOrgRoleAssignmentsArgsForCall(0)
				Expect(orgGUID).To(Equal("org-guid"))

				expectedAssignment := RoleAssignment{ResourceGUID: "org-guid", UserGUID: "user-guid", Roles: []string{"org_auditor"}}
				assignment := <-assignments
				Expect(assignment).To(Equal(expectedAssignment))
			})
		})

		Context("when the capi client call for orgs returns an error", func() {
			BeforeEach(func() {
				client.GetOrgGUIDsReturns([]string{}, fmt.Errorf("org-guids-error"))
			})
			It("sends an error to the errors channel", func() {
				FetchCAPIEntities(client, logger, assignments, errs)
				Expect(client.GetOrgGUIDsCallCount()).To(Equal(1))
				actualError := <-errs
				Expect(actualError).To(MatchError("org-guids-error"))
			})
		})

		Context("when the capi client call for org role assignment returns an error", func() {
			BeforeEach(func() {
				client.GetOrgGUIDsReturns([]string{"org-guid"}, nil)
				client.GetOrgRoleAssignmentsReturns([]RoleAssignment{}, fmt.Errorf("org-role-assignments-error"))
			})
			It("sends an error to the errors channel", func() {
				FetchCAPIEntities(client, logger, assignments, errs)
				Expect(client.GetOrgGUIDsCallCount()).To(Equal(1))
				Expect(client.GetOrgRoleAssignmentsCallCount()).To(Equal(1))
				actualError := <-errs
				Expect(actualError).To(MatchError("org-role-assignments-error"))
				select {
				case <-assignments:
					Fail("assignments channel should be empty")
				default:
				}
			})
		})

		Context("when the capi client returns a space with a space role assignment", func() {
			BeforeEach(func() {
				client.GetSpaceGUIDsReturns([]string{"space-guid"}, nil)
				client.GetSpaceRoleAssignmentsReturns(
					[]RoleAssignment{
						{
							ResourceGUID: "space-guid",
							UserGUID:     "user-guid",
							Roles:        []string{"space_developer"},
						},
					}, nil)

			})
			It("returns a space developer assignment to the channel", func() {
				FetchCAPIEntities(client, logger, assignments, errs)

				Expect(client.GetSpaceGUIDsCallCount()).To(Equal(1))
				_, orgGUID := client.GetSpaceGUIDsArgsForCall(0)
				Expect(orgGUID).To(Equal("org-guid"))
				_, spaceGUID := client.GetSpaceRoleAssignmentsArgsForCall(0)
				Expect(spaceGUID).To(Equal("space-guid"))

				assignment := <-assignments
				expectedAssignment := RoleAssignment{ResourceGUID: "space-guid", UserGUID: "user-guid", Roles: []string{"space_developer"}}
				Expect(assignment).To(Equal(expectedAssignment))

			})
		})

		Context("when the capi client call for spaces returns an error", func() {
			BeforeEach(func() {
				client.GetSpaceGUIDsReturns([]string{}, fmt.Errorf("space-guid-error"))
			})
			It("sends an error to the errors channel", func() {
				FetchCAPIEntities(client, logger, assignments, errs)
				Expect(client.GetSpaceGUIDsCallCount()).To(Equal(1))
				actualError := <-errs
				Expect(actualError).To(MatchError("space-guid-error"))
			})
		})

		Context("when the capi client call for space role assignment returns an error", func() {
			BeforeEach(func() {
				client.GetSpaceGUIDsReturns([]string{"space-guid"}, nil)
				client.GetSpaceRoleAssignmentsReturns([]RoleAssignment{}, fmt.Errorf("space-role-assignment-error"))
			})
			It("sends an error to the errors channel", func() {
				FetchCAPIEntities(client, logger, assignments, errs)
				Expect(client.GetSpaceGUIDsCallCount()).To(Equal(1))
				Expect(client.GetSpaceRoleAssignmentsCallCount()).To(Equal(1))
				actualError := <-errs
				Expect(actualError).To(MatchError("space-role-assignment-error"))
				select {
				case <-assignments:
					Fail("assignments channel should be empty")
				default:
				}
			})

			Context("where the org auditor role assignment was still returned from capi", func() {
				BeforeEach(func() {
					client.GetOrgRoleAssignmentsReturns(
						[]RoleAssignment{
							{
								ResourceGUID: "org-guid",
								UserGUID:     "user-guid",
								Roles:        []string{"org_auditor"},
							},
						}, nil)

				})
				It("returns an org assignment and an error to their channels", func() {
					FetchCAPIEntities(client, logger, assignments, errs)
					Expect(client.GetSpaceGUIDsCallCount()).To(Equal(1))
					Expect(client.GetSpaceRoleAssignmentsCallCount()).To(Equal(1))

					actualError := <-errs
					Expect(actualError).To(MatchError("space-role-assignment-error"))
					assignment := <-assignments
					expectedAssignment := RoleAssignment{ResourceGUID: "org-guid", UserGUID: "user-guid", Roles: []string{"org_auditor"}}
					Expect(assignment).To(Equal(expectedAssignment))
				})

			})
		})
	})
})
