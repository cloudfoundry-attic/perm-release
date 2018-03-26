package migrator

import (
	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
	"code.cloudfoundry.org/lager"
	"io"
	"log"
	"os"
	"sync"
)

//go:generate counterfeiter . Retriever
type Retriever interface {
	FetchCAPIEntities(logger lager.Logger, progress *log.Logger, assignments chan<- models.RoleAssignment, errs chan<- error)
}

//go:generate counterfeiter . Retriever
type Reporter interface {
	GenerateReport(w io.Writer, roleAssignments <-chan models.RoleAssignment, errors <-chan error)
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

func (m *Migrator) Migrate(logger lager.Logger, progress *log.Logger) {
	roleAssignments := make(chan models.RoleAssignment)
	errs := make(chan error)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		m.retriever.FetchCAPIEntities(logger, progress, roleAssignments, errs)
	}()

	go func() {
		defer wg.Done()
		m.reporter.GenerateReport(os.Stderr, roleAssignments, errs)
	}()

	wg.Wait()
}
