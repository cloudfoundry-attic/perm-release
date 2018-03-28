package migrator_test

import (
	"fmt"

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
		populator *migratorfakes.FakePopulator
		reporter  *migratorfakes.FakeReporter

		namespace string

		subject *Migrator
	)

	ginkgo.BeforeEach(func() {
		retriever = new(migratorfakes.FakeRetriever)
		populator = new(migratorfakes.FakePopulator)
		reporter = new(migratorfakes.FakeReporter)

		namespace = "fake-namespace"

		subject = NewMigrator(retriever, populator, reporter, namespace)
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

			ginkgo.It("should populate all orgs and spaces", func() {
				subject.Migrate(logger, progressLogger, buffer, dryRun)

				Expect(populator.PopulateOrganizationCallCount()).To(Equal(len(expectedOrgs)))
				Expect(populator.PopulateSpaceCallCount()).To(Equal(len(expectedSpaces)))

				var (
					populatedOrgs   []models.Organization
					populatedSpaces []models.Space
				)

				for i := 0; i < len(expectedOrgs); i++ {
					_, org, ns := populator.PopulateOrganizationArgsForCall(i)
					populatedOrgs = append(populatedOrgs, org)

					Expect(ns).To(Equal(namespace))
				}

				for i := 0; i < len(expectedSpaces); i++ {
					_, space, ns := populator.PopulateSpaceArgsForCall(i)
					populatedSpaces = append(populatedSpaces, space)

					Expect(ns).To(Equal(namespace))
				}

				for _, org := range expectedOrgs {
					Expect(populatedOrgs).To(ContainElement(org))
				}

				for _, space := range expectedSpaces {
					Expect(populatedSpaces).To(ContainElement(space))
				}
			})

			ginkgo.It("retrieves and reports on the role assignments", func() {
				var (
					orgErrs   []error
					spaceErrs []error
				)

				for i := range expectedOrgs {
					err := errors.New(fmt.Sprintf("populate-organization-err-%d", i))
					orgErrs = append(orgErrs, err)
					populator.PopulateOrganizationReturnsOnCall(i, []error{err})
				}

				for i := range expectedSpaces {
					err := errors.New(fmt.Sprintf("populate-space-err-%d", i))
					spaceErrs = append(spaceErrs, err)
					populator.PopulateSpaceReturnsOnCall(i, []error{err})
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

				expectedErrs = append(expectedErrs, append(orgErrs, spaceErrs...)...)

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

			ginkgo.It("should not populate any orgs or spaces", func() {
				subject.Migrate(logger, progressLogger, buffer, dryRun)

				Expect(populator.PopulateOrganizationCallCount()).To(Equal(0))
				Expect(populator.PopulateSpaceCallCount()).To(Equal(0))
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
