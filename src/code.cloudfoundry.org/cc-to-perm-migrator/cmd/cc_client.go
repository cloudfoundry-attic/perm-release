package cmd

import (
	"io"

	"encoding/json"

	"fmt"

	"code.cloudfoundry.org/cc-to-perm-migrator/cloudcontroller"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter . CloudControllerAPIClient

type CloudControllerAPIClient interface {
	MakePaginatedGetRequest(logger lager.Logger, route string, bodyCallback func(lager.Logger, io.Reader) error) error
	GetOrganizations(logger lager.Logger) ([]cloudcontroller.OrganizationResource, error)
}

func IterateOverCloudControllerEntities(logger lager.Logger, roleAssignments chan<- RoleAssignment, c CloudControllerAPIClient) error {
	var (
		route string
		err   error
	)

	// List Organizations
	route = "/v2/organizations"

	var organizations []cloudcontroller.OrganizationResource

	err = c.MakePaginatedGetRequest(logger, route, func(logger lager.Logger, r io.Reader) error {
		var listOrganizationsResponse cloudcontroller.ListOrganizationsResponse
		err = json.NewDecoder(r).Decode(&listOrganizationsResponse)
		if err != nil {
			logger.Error("failed-to-decode-response", err)
			return nil
		}

		organizations = append(organizations, listOrganizationsResponse.Resources...)
		return nil
	})
	if err != nil {
		logger.Error("failed-to-fetch-organizations", err)
	}

	var spaces []cloudcontroller.SpaceResource

	type RoleRequest struct {
		Route string
		Role  string
	}

	for _, organization := range organizations {
		orgGUID := organization.Metadata.GUID
		route = fmt.Sprintf("/v2/organizations/%s/spaces", orgGUID)

		err = c.MakePaginatedGetRequest(logger, route, func(logger lager.Logger, r io.Reader) error {
			var listOrganizationSpacesResponse cloudcontroller.ListSpacesResponse
			err = json.NewDecoder(r).Decode(&listOrganizationSpacesResponse)
			if err != nil {
				logger.Error("failed-to-decode-response", err)
				return nil
			}

			spaces = append(spaces, listOrganizationSpacesResponse.Resources...)
			return nil
		})
		if err != nil {
			logger.Error("failed-to-fetch-organizations", err)
		}

		var users []cloudcontroller.OrgUserResource

		route = fmt.Sprintf("/v2/organizations/%s/user_roles", orgGUID)

		err = c.MakePaginatedGetRequest(logger, route, func(logger lager.Logger, r io.Reader) error {
			var listUsersResponse cloudcontroller.ListOrganizationRolesResponse
			err = json.NewDecoder(r).Decode(&listUsersResponse)
			if err != nil {
				logger.Error("failed-to-decode-response", err)
				return nil
			}

			users = listUsersResponse.Resources
			for _, u := range users {
				for _, role := range u.Entity.Roles {
					roleAssignments <- RoleAssignment{
						RoleName:     role,
						ResourceGUID: orgGUID,
						UserGUID:     u.Metadata.GUID,
					}
				}
			}

			return nil
		})
		if err != nil {
			logger.Error(fmt.Sprintf("failed-to-fetch-assignments-for-org-%s", orgGUID), err)
		}

	}

	for _, space := range spaces {
		spaceGUID := space.Metadata.GUID

		route = fmt.Sprintf("/v2/spaces/%s/user_roles", spaceGUID)
		var users []cloudcontroller.SpaceUserResource

		err = c.MakePaginatedGetRequest(logger, route, func(logger lager.Logger, r io.Reader) error {
			var listUsersResponse cloudcontroller.ListSpaceRolesResponse
			err = json.NewDecoder(r).Decode(&listUsersResponse)
			if err != nil {
				logger.Error("failed-to-decode-response", err)
				return nil
			}

			users = listUsersResponse.Resources
			for _, u := range users {
				for _, role := range u.Entity.Roles {
					roleAssignments <- RoleAssignment{
						RoleName:     role,
						ResourceGUID: space.Metadata.GUID,
						UserGUID:     u.Metadata.GUID,
					}
				}

			}

			return nil
		})
		if err != nil {
			logger.Error(fmt.Sprintf("failed-to-fetch-assignments-for-space-%s", spaceGUID), err)
		}
	}

	close(roleAssignments)
	return nil
}

type RoleAssignment struct {
	RoleName     string
	ResourceGUID string
	UserGUID     string
}
