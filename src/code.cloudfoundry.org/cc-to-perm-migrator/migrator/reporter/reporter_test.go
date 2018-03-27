package reporter_test

import (
	. "code.cloudfoundry.org/cc-to-perm-migrator/migrator/reporter"

	"errors"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = ginkgo.Describe("Reporter", func() {
	var (
		subject *Reporter

		buffer *Buffer
		errs   []error
	)

	ginkgo.BeforeEach(func() {
		buffer = NewBuffer()
		errs = append(errs, errors.New("There has been a problem"))

		for i := 0; i < 3; i++ {
			errs = append(errs, &models.ErrorEvent{
				Cause:      errors.New("failed-to-fetch-orgs"),
				GUID:       "",
				EntityType: "/v2/organizations",
			})
		}

		for i := 0; i < 4; i++ {
			errs = append(errs, &models.ErrorEvent{
				Cause:      errors.New("failed-to-decode-response"),
				GUID:       "",
				EntityType: "/v2/organizations/org-guid/user_roles",
			})
		}

		for i := 0; i < 2; i++ {
			errs = append(errs, &models.ErrorEvent{
				Cause:      errors.New("something-happened"),
				GUID:       "org-guid",
				EntityType: "/v2/organizations/org-guid/user_roles",
			})
		}
	})

	ginkgo.AfterEach(func() {
		buffer.Close()
	})

	ginkgo.Describe("#GenerateReport", func() {
		ginkgo.It("reports on the number of assignments and errors encountered", func() {
			orgs := []models.Organization{
				{
					GUID: "org-guid-1",
					Assignments: []models.RoleAssignment{
						{
							UserGUID: "user-guid-1",
							Roles:    []string{"org_auditor", "org_manager"},
						},
						{
							UserGUID: "user-guid-2",
							Roles:    []string{"org_auditor"},
						},
					},
				},
				{
					GUID: "org-guid-2",
					Assignments: []models.RoleAssignment{
						{
							UserGUID: "user-guid-1",
							Roles:    []string{"org_manager"},
						},
					},
				},
			}
			spaces := []models.Space{
				{
					GUID:    "space-guid-1",
					OrgGUID: "org-guid-2",
					Assignments: []models.RoleAssignment{
						{
							UserGUID: "user-guid-1",
							Roles:    []string{"space_auditor", "space_manager"},
						},
					},
				},
			}

			subject.GenerateReport(buffer, orgs, spaces, errs)

			Expect(buffer).To(Say("Report\n"))
			Expect(buffer).To(Say("Number of role assignments: 4"))
			Expect(buffer).To(Say("Total errors: 10"))
			Expect(buffer).To(Say("Summary\n"))
			Expect(buffer).To(Say("For /v2/organizations:"))
			Expect(buffer).To(Say("3 failed-to-fetch-orgs"))
			Expect(buffer).To(Say("For /v2/organizations/org-guid/user_roles:"))
			Expect(buffer).To(Say("4 failed-to-decode-response"))
			Expect(buffer).To(Say("2 something-happened"))
			Expect(buffer).To(Say("Other errors"))
			Expect(buffer).To(Say("1 There has been a problem"))
		})
	})
})
