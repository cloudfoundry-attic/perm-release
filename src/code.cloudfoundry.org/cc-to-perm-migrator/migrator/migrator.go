package migrator

import (
	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
	"code.cloudfoundry.org/lager"
	"io"
	"log"
	"sync"
)

//go:generate counterfeiter . Retriever
type Retriever interface {
	FetchRoleAssignments(logger lager.Logger, progressLogger *log.Logger, assignments chan<- models.RoleAssignment, errs chan<- error)
}

//go:generate counterfeiter . Reporter
type Reporter interface {
	GenerateReport(w io.Writer, numAssignments int, errs []error)
}

type Migrator struct {
	retriever Retriever
	reporter  Reporter
}

func NewMigrator(retriever Retriever, reporter Reporter) *Migrator {
	return &Migrator{
		retriever: retriever,
		reporter:  reporter,
	}
}

func (m *Migrator) Migrate(logger lager.Logger, progressLogger *log.Logger, writer io.Writer) {
	assignmentChan := make(chan models.RoleAssignment)
	errChan := make(chan error)

	var (
		count int
		errs  []error
		wg    sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer close(assignmentChan)
		defer close(errChan)
		m.retriever.FetchRoleAssignments(logger, progressLogger, assignmentChan, errChan)
	}()

	go func() {
		defer wg.Done()
		for range assignmentChan {
			count++
		}
	}()

	go func() {
		defer wg.Done()
		for err := range errChan {
			errs = append(errs, err)
		}
	}()

	wg.Wait()

	m.reporter.GenerateReport(writer, count, errs)
}
