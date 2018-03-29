package retriever

import (
	"log"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
	"code.cloudfoundry.org/lager"
	"fmt"
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

func (r *Retriever) FetchResources(logger lager.Logger, progress *log.Logger, orgs chan<- models.Organization, spaces chan<- models.Space, errs chan<- error) {
	orgGUIDs, err := r.client.GetOrgGUIDs(logger)
	if err != nil {
		errs <- err
	}

	progress.Printf("Fetched %d org GUIDs", len(orgGUIDs))

	for orgIndex, orgGUID := range orgGUIDs {
		orgLogPrefix := fmt.Sprintf("[org:%s %d/%d]", orgGUID, orgIndex+1, len(orgGUIDs))
		orgAssignments, err := r.client.GetOrgRoleAssignments(logger, orgGUID)
		if err != nil {
			errs <- err
		}
		progress.Printf("%s Fetched %d org role assignments.", orgLogPrefix, len(orgAssignments))

		orgs <- models.Organization{
			GUID:        orgGUID,
			Assignments: orgAssignments,
		}

		spaceGUIDs, err := r.client.GetSpaceGUIDs(logger, orgGUID)

		if err != nil {
			errs <- err
		}

		for spaceIndex, spaceGUID := range spaceGUIDs {
			orgSpaceLogPrefix := fmt.Sprintf("%s [space:%s %d/%d]", orgLogPrefix, spaceGUID, spaceIndex+1, len(spaceGUIDs))

			spaceAssignments, err := r.client.GetSpaceRoleAssignments(logger, spaceGUID)
			if err != nil {
				errs <- err
			}
			progress.Printf("%s Fetched %d space role assignments.", orgSpaceLogPrefix, len(spaceAssignments))

			spaces <- models.Space{
				GUID:        spaceGUID,
				OrgGUID:     orgGUID,
				Assignments: spaceAssignments,
			}
		}
	}
	progress.Printf("Done.")
}
