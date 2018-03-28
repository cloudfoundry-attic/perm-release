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

//go:generate counterfeiter . Populator

type Populator interface {
	PopulateOrganization(logger lager.Logger, org models.Organization, namespace string) []error
	PopulateSpace(logger lager.Logger, space models.Space, namespace string) []error
}

//go:generate counterfeiter . Reporter

type Reporter interface {
	GenerateReport(w io.Writer, orgs []models.Organization, spaces []models.Space, errs []error)
}

type Migrator struct {
	retriever Retriever
	populator Populator
	reporter  Reporter
	namespace string
}

func NewMigrator(retriever Retriever, populator Populator, reporter Reporter, namespace string) *Migrator {
	return &Migrator{
		retriever: retriever,
		populator: populator,
		reporter:  reporter,
		namespace: namespace,
	}
}

func (m *Migrator) Migrate(logger lager.Logger, progressLogger *log.Logger, writer io.Writer, dryRun bool) {
	orgChan := make(chan models.Organization)
	spaceChan := make(chan models.Space)
	errChan := make(chan error)

	var (
		orgs              []models.Organization
		spaces            []models.Space
		retrieveErrs      []error
		populateOrgErrs   []error
		populateSpaceErrs []error
		wg                sync.WaitGroup
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

		orgLogger := logger.Session("populate-organizations")

		for org := range orgChan {
			orgs = append(orgs, org)

			if !dryRun {
				errs := m.populator.PopulateOrganization(orgLogger, org, m.namespace)
				populateOrgErrs = append(populateOrgErrs, errs...)
			}
		}
	}()

	go func() {
		defer wg.Done()

		spaceLogger := logger.Session("populate-spaces")

		for space := range spaceChan {
			spaces = append(spaces, space)

			if !dryRun {
				errs := m.populator.PopulateSpace(spaceLogger, space, m.namespace)
				populateSpaceErrs = append(populateSpaceErrs, errs...)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for err := range errChan {
			retrieveErrs = append(retrieveErrs, err)
		}
	}()

	wg.Wait()

	errs := append(retrieveErrs, append(populateOrgErrs, populateSpaceErrs...)...)

	m.reporter.GenerateReport(writer, orgs, spaces, errs)
}
