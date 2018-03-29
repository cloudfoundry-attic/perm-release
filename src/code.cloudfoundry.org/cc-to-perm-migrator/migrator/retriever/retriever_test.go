package retriever_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"

	"log"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
	. "code.cloudfoundry.org/cc-to-perm-migrator/migrator/retriever"
	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/retriever/retrieverfakes"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Retriever", func() {
	var (
		orgs           chan models.Organization
		spaces         chan models.Space
		errs           chan error
		client         *retrieverfakes.FakeCAPIClient
		logger         *lagertest.TestLogger
		progressLogger *log.Logger
		progressLog    *gbytes.Buffer
		subject        *Retriever
	)

	BeforeEach(func() {
		orgs = make(chan models.Organization)
		spaces = make(chan models.Space)
		errs = make(chan error)

		client = new(retrieverfakes.FakeCAPIClient)
		logger = lagertest.NewTestLogger("fetch-capi-entities")
		progressLog = gbytes.NewBuffer()
		progressLogger = log.New(progressLog, "", 0)

		subject = NewRetriever(client)
	})

	AfterEach(func() {
		close(orgs)
		close(spaces)
		close(errs)

		progressLog.Close()
	})

	Describe("#FetchResources", func() {
		BeforeEach(func() {
			client.GetOrgGUIDsReturns([]string{"org-guid-1", "org-guid-2"}, nil)
		})

		Context("when the capi client returns orgs with different org role assignment", func() {
			var (
				expectedOrg1 models.Organization
				expectedOrg2 models.Organization
			)

			BeforeEach(func() {
				expectedOrg1 = models.Organization{
					GUID: "org-guid-1",
					Assignments: []models.RoleAssignment{
						{
							UserGUID: "user-guid-1",
							Roles:    []string{"org_auditor", "org_user"},
						},
						{
							UserGUID: "user-guid-2",
							Roles:    []string{"billing_manager"},
						},
					},
				}
				expectedOrg2 = models.Organization{
					GUID: "org-guid-2",
					Assignments: []models.RoleAssignment{
						{
							UserGUID: "user-guid-1",
							Roles:    []string{"org_manager"},
						},
					},
				}

				client.GetOrgRoleAssignmentsReturnsOnCall(0, expectedOrg1.Assignments, nil)
				client.GetOrgRoleAssignmentsReturnsOnCall(1, expectedOrg2.Assignments, nil)
			})

			It("returns the org role assignments to the channel", func(done Done) {
				go subject.FetchResources(logger, progressLogger, orgs, spaces, errs)

				org1, org2 := <-orgs, <-orgs

				Expect(client.GetOrgGUIDsCallCount()).To(Equal(1))
				Expect(client.GetOrgRoleAssignmentsCallCount()).To(Equal(2))

				_, orgGUID := client.GetOrgRoleAssignmentsArgsForCall(0)
				Expect(orgGUID).To(Equal("org-guid-1"))
				_, orgGUID = client.GetOrgRoleAssignmentsArgsForCall(1)
				Expect(orgGUID).To(Equal("org-guid-2"))

				Expect(org1).To(Equal(expectedOrg1))
				Expect(org2).To(Equal(expectedOrg2))

				close(done)
			})

			It("reports the progress of the migration", func(done Done) {
				go subject.FetchResources(logger, progressLogger, orgs, spaces, errs)

				<-orgs
				<-orgs

				Eventually(progressLog).Should(gbytes.Say("Fetched 2 org GUIDs"))
				Eventually(progressLog).Should(gbytes.Say("\\[org:org-guid-1 1/2\\] Fetched 2 org role assignments."))
				Eventually(progressLog).Should(gbytes.Say("\\[org:org-guid-2 2/2\\] Fetched 1 org role assignments."))
				Eventually(progressLog).Should(gbytes.Say("Done."))

				close(done)
			})
		})

		Context("when the capi client call for orgs returns an error", func() {
			BeforeEach(func() {
				client.GetOrgGUIDsReturns([]string{}, fmt.Errorf("org-guids-error"))
			})

			It("sends an error to the errors channel", func(done Done) {
				go subject.FetchResources(logger, progressLogger, orgs, spaces, errs)

				actualErrorEvent := <-errs
				Expect(actualErrorEvent).To(MatchError("org-guids-error"))
				Expect(client.GetOrgGUIDsCallCount()).To(Equal(1))

				close(done)
			})
		})

		Context("when the capi client call for org role assignment returns an error", func() {
			BeforeEach(func() {
				client.GetOrgGUIDsReturns([]string{"org-guid"}, nil)
				client.GetOrgRoleAssignmentsReturns([]models.RoleAssignment{}, fmt.Errorf("org-role-assignments-error"))
			})

			It("sends an error to the errors channel", func(done Done) {
				go subject.FetchResources(logger, progressLogger, orgs, spaces, errs)

				err := <-errs
				<-orgs

				Expect(client.GetOrgGUIDsCallCount()).To(Equal(1))
				Expect(client.GetOrgRoleAssignmentsCallCount()).To(Equal(1))

				Expect(err).To(MatchError("org-role-assignments-error"))

				close(done)
			})
		})

		Context("when the capi client returns spaces with space role assignments", func() {
			var (
				expectedSpace1 models.Space
				expectedSpace2 models.Space
			)

			BeforeEach(func() {
				client.GetSpaceGUIDsReturnsOnCall(0, []string{"space-guid-1"}, nil)
				client.GetSpaceGUIDsReturnsOnCall(1, []string{"space-guid-2"}, nil)

				expectedSpace1 = models.Space{
					GUID:    "space-guid-1",
					OrgGUID: "org-guid-1",
					Assignments: []models.RoleAssignment{
						{
							UserGUID: "user-guid-1",
							Roles:    []string{"space_developer", "space_manager"},
						},
						{
							UserGUID: "user-guid-2",
							Roles:    []string{"space_auditor"},
						},
					},
				}
				expectedSpace2 = models.Space{
					GUID:    "space-guid-2",
					OrgGUID: "org-guid-2",
					Assignments: []models.RoleAssignment{
						{
							UserGUID: "user-guid-4",
							Roles:    []string{"space_manager"},
						},
						{
							UserGUID: "user-guid-5",
							Roles:    []string{"space_developer"},
						},
					},
				}

				client.GetSpaceRoleAssignmentsReturnsOnCall(0, expectedSpace1.Assignments, nil)
				client.GetSpaceRoleAssignmentsReturnsOnCall(1, expectedSpace2.Assignments, nil)
			})

			It("returns the space role assignments to the channel", func(done Done) {
				go subject.FetchResources(logger, progressLogger, orgs, spaces, errs)

				<-orgs
				space1 := <-spaces
				<-orgs
				space2 := <-spaces

				Expect(client.GetSpaceGUIDsCallCount()).To(Equal(2))
				_, orgGUID := client.GetSpaceGUIDsArgsForCall(0)
				Expect(orgGUID).To(Equal("org-guid-1"))
				_, orgGUID = client.GetSpaceGUIDsArgsForCall(1)
				Expect(orgGUID).To(Equal("org-guid-2"))
				_, spaceGUID := client.GetSpaceRoleAssignmentsArgsForCall(0)
				Expect(spaceGUID).To(Equal("space-guid-1"))
				_, spaceGUID = client.GetSpaceRoleAssignmentsArgsForCall(1)
				Expect(spaceGUID).To(Equal("space-guid-2"))

				Expect(space1).To(Equal(expectedSpace1))
				Expect(space2).To(Equal(expectedSpace2))

				close(done)
			})
		})

		Context("when the capi client call for spaces returns an error", func() {
			BeforeEach(func() {
				client.GetOrgGUIDsReturns([]string{"org-guid"}, nil)
				client.GetSpaceGUIDsReturns([]string{}, fmt.Errorf("space-guid-error"))
			})

			It("sends an error to the errors channel", func(done Done) {
				go subject.FetchResources(logger, progressLogger, orgs, spaces, errs)

				<-orgs
				actualErrorEvent := <-errs

				Expect(client.GetSpaceGUIDsCallCount()).To(Equal(1))
				Expect(actualErrorEvent).To(MatchError("space-guid-error"))

				close(done)
			})
		})

		Context("when the capi client call for space role assignment returns an error", func() {
			BeforeEach(func() {
				client.GetOrgGUIDsReturns([]string{"org-guid"}, nil)
				client.GetSpaceGUIDsReturns([]string{"space-guid"}, nil)
				client.GetSpaceRoleAssignmentsReturns([]models.RoleAssignment{}, fmt.Errorf("space-role-assignment-error"))
			})

			It("sends an error to the errors channel", func(done Done) {
				go subject.FetchResources(logger, progressLogger, orgs, spaces, errs)

				<-orgs
				actualErrorEvent := <-errs
				<-spaces

				Expect(client.GetSpaceGUIDsCallCount()).To(Equal(1))
				Expect(client.GetSpaceRoleAssignmentsCallCount()).To(Equal(1))
				Expect(actualErrorEvent).To(MatchError("space-role-assignment-error"))

				close(done)
			})

			Context("where the org auditor role assignment was still returned from capi", func() {
				var expectedOrg models.Organization

				BeforeEach(func() {
					expectedOrg = models.Organization{
						GUID: "org-guid",
						Assignments: []models.RoleAssignment{
							{
								UserGUID: "user-guid",
								Roles:    []string{"org_auditor"},
							},
						},
					}
					client.GetOrgRoleAssignmentsReturns(
						[]models.RoleAssignment{
							{
								UserGUID: "user-guid",
								Roles:    []string{"org_auditor"},
							},
						}, nil)

				})

				It("returns an org assignment and an error to their channels", func(done Done) {
					go subject.FetchResources(logger, progressLogger, orgs, spaces, errs)

					org := <-orgs
					actualErrorEvent := <-errs
					<-spaces

					Expect(client.GetSpaceGUIDsCallCount()).To(Equal(1))
					Expect(client.GetSpaceRoleAssignmentsCallCount()).To(Equal(1))

					Expect(actualErrorEvent).To(MatchError("space-role-assignment-error"))
					Expect(org).To(Equal(expectedOrg))

					close(done)
				})
			})
		})
	})

})
