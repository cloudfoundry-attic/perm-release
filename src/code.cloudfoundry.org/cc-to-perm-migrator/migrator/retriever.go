package migrator

import (
	"code.cloudfoundry.org/lager"
)

type RoleAssignment struct {
	ResourceGUID string
	UserGUID     string
	Roles        []string
}

type ErrorEvent struct {
	Cause error
	GUID string
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

func FetchCAPIEntities(client CAPIClient, logger lager.Logger, assignments chan<- RoleAssignment, errs chan<- error) {
	organizations, err := client.GetOrgGUIDs(logger)
	if err != nil {
		errs <- err
	}

	for _, organization := range organizations {
		orgAssignments, err := client.GetOrgRoleAssignments(logger, organization)
		if err != nil {
			errs <- err
		}

		for _, assignment := range orgAssignments {
			assignments <- assignment
		}

		spaces, err := client.GetSpaceGUIDs(logger, organization)
		if err != nil {
			errs <- err
		}

		for _, space := range spaces {
			spaceAssignments, err := client.GetSpaceRoleAssignments(nil, space)
			if err != nil {
				errs <- err
			}

			for _, assignment := range spaceAssignments {
				assignments <- assignment
			}
		}
	}

	close(assignments)
	close(errs)
}
