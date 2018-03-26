package retriever

import (
	"log"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter . CAPIClient

type CAPIClient interface {
	GetOrgGUIDs(logger lager.Logger) ([]string, error)
	GetSpaceGUIDs(logger lager.Logger, orgGUID string) ([]string, error)
	GetOrgRoleAssignments(logger lager.Logger, orgGUID string) ([]models.RoleAssignment, error)
	GetSpaceRoleAssignments(logger lager.Logger, spaceGUID string) ([]models.RoleAssignment, error)
}

type Retriever struct {
	client CAPIClient
}

func NewRetriever(client CAPIClient) *Retriever {
	return &Retriever{
		client: client,
	}
}

func (r *Retriever) FetchRoleAssignments(logger lager.Logger, progress *log.Logger, assignments chan<- models.RoleAssignment, errs chan<- error) {
	organizations, err := r.client.GetOrgGUIDs(logger)
	if err != nil {
		errs <- err
	}
	progress.Printf("Fetched %d org GUIDs", len(organizations))
	for orgIndex, organization := range organizations {
		progress.Printf("Processing org %s [%d/%d]", organization, orgIndex+1, len(organizations))
		orgAssignments, err := r.client.GetOrgRoleAssignments(logger, organization)
		if err != nil {
			errs <- err
		}
		progress.Printf("%s: Fetched %d org role assignments. Migrating...", organization, len(orgAssignments))

		for _, assignment := range orgAssignments {
			assignments <- assignment
		}

		spaces, err := r.client.GetSpaceGUIDs(logger, organization)
		if err != nil {
			errs <- err
		}
		progress.Printf("%s: Fetched %d spaces. Migrating...", organization, len(orgAssignments))
		for _, space := range spaces {
			spaceAssignments, err := r.client.GetSpaceRoleAssignments(logger, space)
			if err != nil {
				errs <- err
			}

			for _, assignment := range spaceAssignments {
				assignments <- assignment
			}
		}
	}
	progress.Printf("Done.")
}
