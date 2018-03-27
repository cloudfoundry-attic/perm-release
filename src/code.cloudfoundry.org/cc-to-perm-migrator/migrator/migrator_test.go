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

		logger         *lagertest.TestLogger
		progressLogger *log.Logger

		buffer *Buffer

		subject *Migrator
	)

	ginkgo.BeforeEach(func() {
		retriever = new(migratorfakes.FakeRetriever)
		reporter = new(migratorfakes.FakeReporter)

		buffer = NewBuffer()

		logger = lagertest.NewTestLogger("migrator")
		progressLogger = log.New(buffer, "", 0)

		subject = NewMigrator(retriever, reporter)
	})

	ginkgo.AfterEach(func() {
		buffer.Close()
	})

	ginkgo.It("retrieves and reports on the role assignments", func() {
		expectedOrgAssignments := []models.RoleAssignment{
			{
				UserGUID: "user-guid",
				Roles:    []string{"org_auditor"},
			},
			{
				UserGUID: "user-guid",
				Roles:    []string{"org_user"},
			},
		}
		expectedOrgs := []models.Organization{
			{
				GUID:        "org-guid",
				Assignments: expectedOrgAssignments,
			},
		}

		expectedSpaceAssignments := []models.RoleAssignment{
			{
				UserGUID: "user-guid",
				Roles:    []string{"space_developer"},
			},
			{
				UserGUID: "user-guid",
				Roles:    []string{"space_manager"},
			},
		}
		expectedSpaces := []models.Space{
			{
				GUID:        "space-guid",
				Assignments: expectedSpaceAssignments,
			},
		}

		expectedErrs := []error{
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

		subject.Migrate(logger, progressLogger, buffer)

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
