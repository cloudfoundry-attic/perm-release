package migrator

import (
	"io"
	"log"
	"sync"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter . Retriever

type Retriever interface {
	FetchResources(logger lager.Logger, progressLogger *log.Logger, orgs chan<- models.Organization, spaces chan<- models.Space, errs chan<- error)
}

//go:generate counterfeiter . Reporter

type Reporter interface {
	GenerateReport(w io.Writer, orgs []models.Organization, spaces []models.Space, errs []error)
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
		orgs   []models.Organization
		spaces []models.Space
		errs   []error
		wg     sync.WaitGroup
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
			orgs = append(orgs, org)
		}
	}()

	go func() {
		defer wg.Done()
		for space := range spaceChan {
			spaces = append(spaces, space)
		}
	}()

	go func() {
		defer wg.Done()
		for err := range errChan {
			errs = append(errs, err)
		}
	}()

	wg.Wait()

	m.reporter.GenerateReport(writer, orgs, spaces, errs)
}
