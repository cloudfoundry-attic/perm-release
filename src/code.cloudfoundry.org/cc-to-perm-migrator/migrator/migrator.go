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
	FetchResources(logger lager.Logger, progressLogger *log.Logger, orgs chan<- models.Organization, spaces chan<- models.Space, errs chan<- error)
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
	orgChan := make(chan models.Organization)
	spaceChan := make(chan models.Space)
	errChan := make(chan error)

	var (
		count int
		errs  []error
		wg    sync.WaitGroup
	)

	wg.Add(3)

	go func() {
		defer close(orgChan)
		defer close(spaceChan)
		defer close(errChan)
		m.retriever.FetchResources(logger, progressLogger, orgChan, spaceChan, errChan)
	}()

	go func() {
		defer wg.Done()
		for org := range orgChan {
			for range org.Assignments {
				count++
			}
		}
	}()

	go func() {
		defer wg.Done()
		for space := range spaceChan {
			for range space.Assignments {
				count++
			}
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
