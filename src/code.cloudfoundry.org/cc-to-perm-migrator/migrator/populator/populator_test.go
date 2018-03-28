package populator_test

import (
	"errors"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
	. "code.cloudfoundry.org/cc-to-perm-migrator/migrator/populator"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm-go"
	permgofakes "code.cloudfoundry.org/perm-go/perm-gofakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Populator", func() {
	var (
		client *permgofakes.FakeRoleServiceClient
		logger *lagertest.TestLogger

		namespace string

		subject *Populator
	)

	BeforeEach(func() {
		namespace = "fake-namespace"

		client = new(permgofakes.FakeRoleServiceClient)
		logger = lagertest.NewTestLogger("populator")

		subject = NewPopulator(client)
	})

	Describe("#PopulateOrganization", func() {
		var (
			org models.Organization
		)

		BeforeEach(func() {
			org = models.Organization{
				GUID: "fake-org-guid",
				Assignments: []models.RoleAssignment{
					{
						UserGUID: "user-guid-1",
						Roles:    []string{"org_auditor", "org_manager"},
					},
					{
						UserGUID: "user-guid-2",
						Roles:    []string{"org_auditor"},
					},
					{
						UserGUID: "user-guid-3",
						Roles:    []string{"org_manager", "org_user"},
					},
				},
			}
		})

		Context("when all requests succeed", func() {
			It("returns no errors", func() {
				errs := subject.PopulateOrganization(logger, org, namespace)

				Expect(errs).To(BeEmpty())
			})

			populatesOrg(func() (*Populator, *permgofakes.FakeRoleServiceClient, lager.Logger, models.Organization, string) {
				return subject, client, logger, org, namespace
			})
		})

		Context("when it fails to create any roles", func() {
			var (
				expectedErrs []error
			)

			BeforeEach(func() {
				err1 := errors.New("fake-create-role-err-1")
				err2 := errors.New("fake-create-role-err-2")

				expectedErrs = []error{err1, err2}

				client.CreateRoleReturnsOnCall(1, nil, err1)
				client.CreateRoleReturnsOnCall(2, nil, err2)
			})

			It("returns the errors", func() {
				errs := subject.PopulateOrganization(logger, org, namespace)

				Expect(errs).To(HaveLen(len(expectedErrs)))

				for _, err := range expectedErrs {
					Expect(errs).To(ContainElement(err))
				}
			})
		})

		Context("when it fails to assign any roles", func() {
			var (
				expectedErrs []error
			)

			BeforeEach(func() {
				err1 := errors.New("fake-assign-role-err-1")
				err2 := errors.New("fake-assign-role-err-2")
				err3 := errors.New("fake-assign-role-err-3")

				expectedErrs = []error{err1, err2, err3}

				client.AssignRoleReturnsOnCall(1, nil, err1)
				client.AssignRoleReturnsOnCall(2, nil, err2)
				client.AssignRoleReturnsOnCall(3, nil, err3)
			})

			It("returns the errors", func() {
				errs := subject.PopulateOrganization(logger, org, namespace)

				Expect(errs).To(HaveLen(len(expectedErrs)))

				for _, err := range expectedErrs {
					Expect(errs).To(ContainElement(err))
				}
			})
		})
	})

	Describe("#PopulateSpace", func() {
		var (
			space models.Space
		)

		BeforeEach(func() {
			space = models.Space{
				GUID:    "fake-space-guid",
				OrgGUID: "fake-org-guid",
				Assignments: []models.RoleAssignment{
					{
						UserGUID: "user-guid-1",
						Roles:    []string{"space_developer", "space_manager"},
					},
					{
						UserGUID: "user-guid-2",
						Roles:    []string{"space_developer"},
					},
					{
						UserGUID: "user-guid-3",
						Roles:    []string{"space_manager", "space_auditor"},
					},
				},
			}
		})

		Context("when all requests succeed", func() {
			It("returns no errors", func() {
				errs := subject.PopulateSpace(logger, space, namespace)

				Expect(errs).To(BeEmpty())
			})

			populatesSpace(func() (*Populator, *permgofakes.FakeRoleServiceClient, lager.Logger, models.Space, string) {
				return subject, client, logger, space, namespace
			})
		})

		Context("when it fails to create any roles", func() {
			var (
				expectedErrs []error
			)

			BeforeEach(func() {
				err1 := errors.New("fake-create-role-err-1")
				err2 := errors.New("fake-create-role-err-2")

				expectedErrs = []error{err1, err2}

				client.CreateRoleReturnsOnCall(1, nil, err1)
				client.CreateRoleReturnsOnCall(2, nil, err2)
			})

			It("returns the errors", func() {
				errs := subject.PopulateSpace(logger, space, namespace)

				Expect(errs).To(HaveLen(len(expectedErrs)))

				for _, err := range expectedErrs {
					Expect(errs).To(ContainElement(err))
				}
			})
		})

		Context("when it fails to assign any roles", func() {
			var (
				expectedErrs []error
			)

			BeforeEach(func() {
				err1 := errors.New("fake-assign-role-err-1")
				err2 := errors.New("fake-assign-role-err-2")
				err3 := errors.New("fake-assign-role-err-3")

				expectedErrs = []error{err1, err2, err3}

				client.AssignRoleReturnsOnCall(1, nil, err1)
				client.AssignRoleReturnsOnCall(2, nil, err2)
				client.AssignRoleReturnsOnCall(3, nil, err3)
			})

			It("returns the errors", func() {
				errs := subject.PopulateSpace(logger, space, namespace)

				Expect(errs).To(HaveLen(len(expectedErrs)))

				for _, err := range expectedErrs {
					Expect(errs).To(ContainElement(err))
				}
			})
		})
	})
})

func populatesOrg(injector func() (*Populator, *permgofakes.FakeRoleServiceClient, lager.Logger, models.Organization, string)) {
	It("creates the org roles", func() {
		subject, client, logger, org, namespace := injector()

		subject.PopulateOrganization(logger, org, namespace)

		roles := []string{"user", "manager", "billing_manager", "auditor"}

		var expectedReqs []*protos.CreateRoleRequest

		for _, role := range roles {
			req := &protos.CreateRoleRequest{
				Name: fmt.Sprintf("org-%s-%s", role, org.GUID),
				Permissions: []*protos.Permission{
					{
						Name:            fmt.Sprintf("org.%s", role),
						ResourcePattern: org.GUID,
					},
				},
			}

			expectedReqs = append(expectedReqs, req)
		}

		var reqs []*protos.CreateRoleRequest

		Expect(client.CreateRoleCallCount()).To(Equal(len(roles)))

		for i := 0; i < len(roles); i++ {
			_, req, _ := client.CreateRoleArgsForCall(i)
			reqs = append(reqs, req)
		}

		for _, req := range expectedReqs {
			Expect(reqs).To(ContainElement(req))
		}
	})

	It("assigns all members to all of their roles", func() {
		subject, client, logger, org, namespace := injector()

		subject.PopulateOrganization(logger, org, namespace)

		var expectedReqs []*protos.AssignRoleRequest

		for _, assignment := range org.Assignments {
			for _, role := range assignment.Roles {
				role = strings.Replace(role, "org_", "", -1)
				role = strings.Replace(role, "org-", "", -1)

				req := &protos.AssignRoleRequest{
					Actor: &protos.Actor{
						ID:     assignment.UserGUID,
						Issuer: namespace,
					},
					RoleName: fmt.Sprintf("org-%s-%s", role, org.GUID),
				}

				expectedReqs = append(expectedReqs, req)
			}
		}

		Expect(client.AssignRoleCallCount()).To(Equal(len(expectedReqs)))

		var reqs []*protos.AssignRoleRequest

		for i := 0; i < len(expectedReqs); i++ {
			_, req, _ := client.AssignRoleArgsForCall(i)
			reqs = append(reqs, req)
		}

		for _, req := range expectedReqs {
			Expect(reqs).To(ContainElement(req))
		}
	})
}

func populatesSpace(injector func() (*Populator, *permgofakes.FakeRoleServiceClient, lager.Logger, models.Space, string)) {
	It("creates the space roles", func() {
		subject, client, logger, space, namespace := injector()

		subject.PopulateSpace(logger, space, namespace)

		roles := []string{"manager", "developer", "auditor"}

		var expectedReqs []*protos.CreateRoleRequest

		for _, role := range roles {
			req := &protos.CreateRoleRequest{
				Name: fmt.Sprintf("space-%s-%s", role, space.GUID),
				Permissions: []*protos.Permission{
					{
						Name:            fmt.Sprintf("space.%s", role),
						ResourcePattern: space.GUID,
					},
				},
			}

			expectedReqs = append(expectedReqs, req)
		}

		var reqs []*protos.CreateRoleRequest

		Expect(client.CreateRoleCallCount()).To(Equal(len(roles)))

		for i := 0; i < len(roles); i++ {
			_, req, _ := client.CreateRoleArgsForCall(i)
			reqs = append(reqs, req)
		}

		for _, req := range expectedReqs {
			Expect(reqs).To(ContainElement(req))
		}
	})

	It("assigns all members to all of their roles", func() {
		subject, client, logger, space, namespace := injector()

		subject.PopulateSpace(logger, space, namespace)

		var expectedReqs []*protos.AssignRoleRequest

		for _, assignment := range space.Assignments {
			for _, role := range assignment.Roles {
				role = strings.Replace(role, "space_", "", -1)
				role = strings.Replace(role, "space-", "", -1)

				req := &protos.AssignRoleRequest{
					Actor: &protos.Actor{
						ID:     assignment.UserGUID,
						Issuer: namespace,
					},
					RoleName: fmt.Sprintf("space-%s-%s", role, space.GUID),
				}

				expectedReqs = append(expectedReqs, req)
			}
		}

		Expect(client.AssignRoleCallCount()).To(Equal(len(expectedReqs)))

		var reqs []*protos.AssignRoleRequest

		for i := 0; i < len(expectedReqs); i++ {
			_, req, _ := client.AssignRoleArgsForCall(i)
			reqs = append(reqs, req)
		}

		for _, req := range expectedReqs {
			Expect(reqs).To(ContainElement(req))
		}
	})
}
