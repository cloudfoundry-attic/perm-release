package migrator_test

import (
	. "code.cloudfoundry.org/cc-to-perm-migrator/migrator"

	"errors"
	"log"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/migratorfakes"
	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = ginkgo.Describe("Migrator", func() {
	var (
		retriever *migratorfakes.FakeRetriever
		reporter  *migratorfakes.FakeReporter

		subject *Migrator
	)

	ginkgo.BeforeEach(func() {
		retriever = new(migratorfakes.FakeRetriever)
		reporter = new(migratorfakes.FakeReporter)

		subject = NewMigrator(retriever, reporter)
	})

	ginkgo.Describe("#Migrate", func() {
		var (
			logger         *lagertest.TestLogger
			progressLogger *log.Logger

			buffer *Buffer

			org1         models.Organization
			org2         models.Organization
			expectedOrgs []models.Organization

			space1         models.Space
			expectedSpaces []models.Space

			expectedErrs []error

			dryRun bool
		)

		ginkgo.BeforeEach(func() {
			buffer = NewBuffer()

			logger = lagertest.NewTestLogger("migrator")
			progressLogger = log.New(buffer, "", 0)

			org1 = models.Organization{
				GUID: "org-guid-1",
				Assignments: []models.RoleAssignment{
					{
						UserGUID: "user-guid",
						Roles:    []string{"org_auditor"},
					},
					{
						UserGUID: "user-guid",
						Roles:    []string{"org_user"},
					},
				},
			}
			org2 = models.Organization{
				GUID: "org-guid-2",
				Assignments: []models.RoleAssignment{
					{
						UserGUID: "user-guid-2",
						Roles:    []string{"org_manager"},
					},
				},
			}
			expectedOrgs = []models.Organization{org1, org2}

			space1 = models.Space{
				GUID: "space-guid",
				Assignments: []models.RoleAssignment{
					{
						UserGUID: "user-guid",
						Roles:    []string{"space_developer"},
					},
					{
						UserGUID: "user-guid",
						Roles:    []string{"space_manager"},
					},
				},
			}
			expectedSpaces = []models.Space{space1}

			expectedErrs = []error{
				errors.New("retrieve-error"),
				errors.New("retrieve-error2"),
			}

			retriever.FetchResourcesStub = func(logger lager.Logger, progressLogger *log.Logger, orgsChan chan<- models.Organization, spacesChan chan<- models.Space, errChan chan<- error) {
				for _, org := range expectedOrgs {
					orgsChan <- org
				}
				for _, space := range expectedSpaces {
					spacesChan <- space
				}
				for _, err := range expectedErrs {
					errChan <- err
				}
			}
		})

		ginkgo.AfterEach(func() {
			buffer.Close()
		})

		ginkgo.Context("in regular (non-dry-run) mode", func() {
			ginkgo.BeforeEach(func() {
				dryRun = false
			})

			ginkgo.It("retrieves and reports on the role assignments", func() {
				subject.Migrate(logger, progressLogger, buffer, dryRun)

				Expect(reporter.GenerateReportCallCount()).To(Equal(1))
				buf, orgs, spaces, errs := reporter.GenerateReportArgsForCall(0)

				Expect(buf).To(Equal(buffer))

				Expect(orgs).To(HaveLen(len(expectedOrgs)))
				for _, org := range expectedOrgs {
					Expect(orgs).To(ContainElement(org))
				}

				Expect(spaces).To(HaveLen(len(expectedSpaces)))
				for _, space := range expectedSpaces {
					Expect(spaces).To(ContainElement(space))
				}

				Expect(errs).To(HaveLen(len(expectedErrs)))
				for _, err := range expectedErrs {
					Expect(errs).To(ContainElement(err))
				}
			})
		})

		ginkgo.Context("in dry-run mode", func() {
			ginkgo.BeforeEach(func() {
				dryRun = true
			})

			ginkgo.It("retrieves and reports on the role assignments", func() {
				expectedErrs := []error{
					errors.New("retrieve-error"),
					errors.New("retrieve-error2"),
				}

				subject.Migrate(logger, progressLogger, buffer, dryRun)

				Expect(reporter.GenerateReportCallCount()).To(Equal(1))
				buf, orgs, spaces, errs := reporter.GenerateReportArgsForCall(0)

				Expect(buf).To(Equal(buffer))

				Expect(orgs).To(HaveLen(len(expectedOrgs)))
				for _, org := range expectedOrgs {
					Expect(orgs).To(ContainElement(org))
				}

				Expect(spaces).To(HaveLen(len(expectedSpaces)))
				for _, space := range expectedSpaces {
					Expect(spaces).To(ContainElement(space))
				}

				Expect(errs).To(HaveLen(len(expectedErrs)))
				for _, err := range expectedErrs {
					Expect(errs).To(ContainElement(err))
				}
			})
		})
	})
})
