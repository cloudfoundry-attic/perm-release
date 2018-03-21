package migrator

import (
	"log"

	"code.cloudfoundry.org/lager"
)

type RoleAssignment struct {
	ResourceGUID string
	UserGUID     string
	Roles        []string
}

type ErrorEvent struct {
	Cause      error
	GUID       string
	EntityType string
}

func (e *ErrorEvent) Error() string {
	return e.Cause.Error()
}

//go:generate counterfeiter . CAPIClient

type CAPIClient interface {
	GetOrgGUIDs(logger lager.Logger) ([]string, error)
	GetSpaceGUIDs(logger lager.Logger, orgGUID string) ([]string, error)
	GetOrgRoleAssignments(logger lager.Logger, orgGUID string) ([]RoleAssignment, error)
	GetSpaceRoleAssignments(logger lager.Logger, spaceGUID string) ([]RoleAssignment, error)
}

func FetchCAPIEntities(client CAPIClient, logger lager.Logger, progress *log.Logger, assignments chan<- RoleAssignment, errs chan<- error) {
	organizations, err := client.GetOrgGUIDs(logger)
	if err != nil {
		errs <- err
	}
	progress.Printf("Fetched %d org GUIDs", len(organizations))
	for orgIndex, organization := range organizations {
		progress.Printf("Processing org %s [%d/%d]", organization, orgIndex+1, len(organizations))
		orgAssignments, err := client.GetOrgRoleAssignments(logger, organization)
		if err != nil {
			errs <- err
		}
		progress.Printf("%s: Fetched %d org role assignments. Migrating...", organization, len(orgAssignments))

		for _, assignment := range orgAssignments {
			assignments <- assignment
		}

		spaces, err := client.GetSpaceGUIDs(logger, organization)
		if err != nil {
			errs <- err
		}
		progress.Printf("%s: Fetched %d spaces. Migrating...", organization, len(orgAssignments))
		for _, space := range spaces {
			spaceAssignments, err := client.GetSpaceRoleAssignments(logger, space)
			if err != nil {
				errs <- err
			}

			for _, assignment := range spaceAssignments {
				assignments <- assignment
			}
		}
	}
	progress.Printf("Done.")
	close(assignments)
	close(errs)
}
