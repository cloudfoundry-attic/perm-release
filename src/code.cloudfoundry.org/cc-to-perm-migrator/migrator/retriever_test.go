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
			client.GetOrgGUIDsReturns([]string{"org-guid-1", "org-guid-2"}, nil)
		})

		Context("when the capi client returns orgs with different org role assignment", func() {
			BeforeEach(func() {
				client.GetOrgRoleAssignmentsReturnsOnCall(0,
					[]RoleAssignment{
						{
							ResourceGUID: "org-guid-1",
							UserGUID:     "user-guid-1",
							Roles:        []string{"org_auditor", "org_user"},
						},
						{
							ResourceGUID: "org-guid-1",
							UserGUID:     "user-guid-2",
							Roles:        []string{"billing_manager"},
						},
					}, nil)
				client.GetOrgRoleAssignmentsReturnsOnCall(1,
					[]RoleAssignment{
						{
							ResourceGUID: "org-guid-2",
							UserGUID:     "user-guid-1",
							Roles:        []string{"org_manager"},
						},
					}, nil)

			})
			It("returns the org role assignments to the channel", func() {
				FetchCAPIEntities(client, logger, assignments, errs)

				Expect(client.GetOrgGUIDsCallCount()).To(Equal(1))
				Expect(client.GetOrgRoleAssignmentsCallCount()).To(Equal(2))
				_, orgGUID := client.GetOrgRoleAssignmentsArgsForCall(0)
				Expect(orgGUID).To(Equal("org-guid-1"))
				_, orgGUID = client.GetOrgRoleAssignmentsArgsForCall(1)
				Expect(orgGUID).To(Equal("org-guid-2"))

				expectedAssignment1 := RoleAssignment{ResourceGUID: "org-guid-1", UserGUID: "user-guid-1", Roles: []string{"org_auditor", "org_user"}}
				expectedAssignment2 := RoleAssignment{ResourceGUID: "org-guid-1", UserGUID: "user-guid-2", Roles: []string{"billing_manager"}}
				expectedAssignment3 := RoleAssignment{ResourceGUID: "org-guid-2", UserGUID: "user-guid-1", Roles: []string{"org_manager"}}
				assignment1, assignment2, assignment3 := <-assignments, <-assignments, <-assignments
				Expect(assignment1).To(Equal(expectedAssignment1))
				Expect(assignment2).To(Equal(expectedAssignment2))
				Expect(assignment3).To(Equal(expectedAssignment3))
				Eventually(assignments).Should(BeClosed())
				Eventually(errs).Should(BeClosed())
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
				Eventually(assignments).Should(BeClosed())
			})
		})

		Context("when the capi client returns spaces with space role assignments", func() {
			BeforeEach(func() {
				client.GetSpaceGUIDsReturnsOnCall(0, []string{"space-guid-1"}, nil)
				client.GetSpaceGUIDsReturnsOnCall(1, []string{"space-guid-2"}, nil)
				client.GetSpaceRoleAssignmentsReturnsOnCall(0,
					[]RoleAssignment{
						{
							ResourceGUID: "space-guid-1",
							UserGUID:     "user-guid-1",
							Roles:        []string{"space_developer", "space_manager"},
						},
						{
							ResourceGUID: "space-guid-1",
							UserGUID:     "user-guid-2",
							Roles:        []string{"space_auditor"},
						},
					}, nil)
				client.GetSpaceRoleAssignmentsReturnsOnCall(1,
					[]RoleAssignment{
						{
							ResourceGUID: "space-guid-2",
							UserGUID:     "user-guid-4",
							Roles:        []string{"space_manager"},
						},
						{
							ResourceGUID: "space-guid-2",
							UserGUID:     "user-guid-5",
							Roles:        []string{"space_developer"},
						},
					}, nil)

			})
			It("returns the space role assignments to the channel", func() {
				FetchCAPIEntities(client, logger, assignments, errs)

				Expect(client.GetSpaceGUIDsCallCount()).To(Equal(2))
				_, orgGUID := client.GetSpaceGUIDsArgsForCall(0)
				Expect(orgGUID).To(Equal("org-guid-1"))
				_, orgGUID = client.GetSpaceGUIDsArgsForCall(1)
				Expect(orgGUID).To(Equal("org-guid-2"))
				_, spaceGUID := client.GetSpaceRoleAssignmentsArgsForCall(0)
				Expect(spaceGUID).To(Equal("space-guid-1"))
				_, spaceGUID = client.GetSpaceRoleAssignmentsArgsForCall(1)
				Expect(spaceGUID).To(Equal("space-guid-2"))

				assignment1, assignment2 := <-assignments, <-assignments
				expectedAssignment1 := RoleAssignment{ResourceGUID: "space-guid-1", UserGUID: "user-guid-1", Roles: []string{"space_developer", "space_manager"}}
				expectedAssignment2 := RoleAssignment{ResourceGUID: "space-guid-1", UserGUID: "user-guid-2", Roles: []string{"space_auditor"}}
				Expect(assignment1).To(Equal(expectedAssignment1))
				Expect(assignment2).To(Equal(expectedAssignment2))

				assignment3, assignment4 := <-assignments, <-assignments
				expectedAssignment3 := RoleAssignment{ResourceGUID: "space-guid-2", UserGUID: "user-guid-4", Roles: []string{"space_manager"}}
				expectedAssignment4 := RoleAssignment{ResourceGUID: "space-guid-2", UserGUID: "user-guid-5", Roles: []string{"space_developer"}}
				Expect(assignment3).To(Equal(expectedAssignment3))
				Expect(assignment4).To(Equal(expectedAssignment4))

			})
		})

		Context("when the capi client call for spaces returns an error", func() {
			BeforeEach(func() {
				client.GetOrgGUIDsReturns([]string{"org-guid"}, nil)
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
				client.GetOrgGUIDsReturns([]string{"org-guid"}, nil)
				client.GetSpaceGUIDsReturns([]string{"space-guid"}, nil)
				client.GetSpaceRoleAssignmentsReturns([]RoleAssignment{}, fmt.Errorf("space-role-assignment-error"))
			})
			It("sends an error to the errors channel", func() {
				FetchCAPIEntities(client, logger, assignments, errs)
				Expect(client.GetSpaceGUIDsCallCount()).To(Equal(1))
				Expect(client.GetSpaceRoleAssignmentsCallCount()).To(Equal(1))
				actualError := <-errs
				Expect(actualError).To(MatchError("space-role-assignment-error"))
				Eventually(assignments).Should(BeClosed())
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
