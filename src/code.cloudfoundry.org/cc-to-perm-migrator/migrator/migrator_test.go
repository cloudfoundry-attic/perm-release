package migrator_test

import (
	. "code.cloudfoundry.org/cc-to-perm-migrator/migrator"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/migratorfakes"
	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"errors"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"log"
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
		expectedAssignments := []models.RoleAssignment{
			{
				ResourceGUID: "resource-guid",
				UserGUID:     "user-guid",
				Roles:        []string{"org_auditor"},
			},
			{
				ResourceGUID: "resource-guid-2",
				UserGUID:     "user-guid",
				Roles:        []string{"org_user"},
			},
		}
		expectedErrs := []error{
			errors.New("retrieve-error"),
			errors.New("retrieve-error2"),
		}

		retriever.FetchRoleAssignmentsStub = func(logger lager.Logger, progressLogger *log.Logger, assignmentChan chan<- models.RoleAssignment, errChan chan<- error) {
			for _, assignment := range expectedAssignments {
				assignmentChan <- assignment
			}
			for _, err := range expectedErrs {
				errChan <- err
			}
		}

		subject.Migrate(logger, progressLogger, buffer)

		Expect(reporter.GenerateReportCallCount()).To(Equal(1))
		buf, numAssignments, errs := reporter.GenerateReportArgsForCall(0)

		Expect(buf).To(Equal(buffer))

		Expect(numAssignments).To(Equal(len(expectedAssignments)))

		Expect(errs).To(HaveLen(len(expectedErrs)))
		for _, err := range expectedErrs {
			Expect(errs).To(ContainElement(err))
		}
	})
})
