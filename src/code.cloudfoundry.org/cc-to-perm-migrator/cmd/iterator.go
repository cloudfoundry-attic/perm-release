package cmd

import (
	"context"
	"io"

	"encoding/json"

	"code.cloudfoundry.org/cc-to-perm-migrator/cloudcontroller"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter . CloudControllerAPIClient

type CloudControllerAPIClient interface {
	MakePaginatedGetRequest(ctx context.Context, logger lager.Logger, route string, bodyCallback func(context.Context, lager.Logger, io.Reader) error) error
}

func IterateOverCloudControllerEntities(ctx context.Context, logger lager.Logger, roleAssignments chan<- RoleAssignment, c CloudControllerAPIClient) error {
	var (
		route string
		err   error
	)

	// List Organizations
	route = "/v2/organizations"

	var organizations []cloudcontroller.OrganizationResource

	err = c.MakePaginatedGetRequest(ctx, logger, route, func(ctx context.Context, logger lager.Logger, r io.Reader) error {
		var listOrganizationsResponse cloudcontroller.ListOrganizationsResponse
		err = json.NewDecoder(r).Decode(&listOrganizationsResponse)
		if err != nil {
			logger.Error("failed-to-decode-response", err)
			return err
		}

		organizations = append(organizations, listOrganizationsResponse.Resources...)
		return nil
	})
	if err != nil {
		return err
	}

	var spaces []cloudcontroller.SpaceResource

	type RoleRequest struct {
		Route string
		Role  string
	}

	for _, organization := range organizations {
		route = organization.Entity.SpacesURL

		err = c.MakePaginatedGetRequest(ctx, logger, route, func(ctx context.Context, logger lager.Logger, r io.Reader) error {
			var listOrganizationSpacesResponse cloudcontroller.ListSpacesResponse
			err = json.NewDecoder(r).Decode(&listOrganizationSpacesResponse)
			if err != nil {
				logger.Error("failed-to-decode-response", err)
				return err
			}

			spaces = append(spaces, listOrganizationSpacesResponse.Resources...)
			return nil
		})
		if err != nil {
			return err
		}

		roleRequests := []RoleRequest{
			{Route: organization.Entity.UsersURL, Role: "org-user"},
			{Route: organization.Entity.BillingManagersURL, Role: "org-billing-manager"},
			{Route: organization.Entity.ManagersURL, Role: "org-manager"},
			{Route: organization.Entity.AuditorsURL, Role: "org-auditor"},
		}

		var users []cloudcontroller.UserResource

		for _, roleRequest := range roleRequests {
			route = roleRequest.Route

			err = c.MakePaginatedGetRequest(ctx, logger, route, func(ctx context.Context, logger lager.Logger, r io.Reader) error {
				var listUsersResponse cloudcontroller.ListUsersResponse
				err = json.NewDecoder(r).Decode(&listUsersResponse)
				if err != nil {
					logger.Error("failed-to-decode-response", err)
					return err
				}

				users = listUsersResponse.Resources
				for _, u := range users {
					roleAssignments <- RoleAssignment{
						RoleName:     roleRequest.Role,
						ResourceGUID: organization.Metadata.GUID,
						UserGUID:     u.Metadata.GUID,
					}
				}

				return nil
			})
			if err != nil {
				return err
			}
		}

	}

	for _, space := range spaces {
		roleRequests := []RoleRequest{
			{Route: space.Entity.DevelopersURL, Role: "space-developer"},
			{Route: space.Entity.AuditorsURL, Role: "space-auditor"},
			{Route: space.Entity.ManagersURL, Role: "space-manager"},
		}

		var users []cloudcontroller.UserResource

		for _, roleRequest := range roleRequests {
			route = roleRequest.Route

			err = c.MakePaginatedGetRequest(ctx, logger, route, func(ctx context.Context, logger lager.Logger, r io.Reader) error {
				var listUsersResponse cloudcontroller.ListUsersResponse
				err = json.NewDecoder(r).Decode(&listUsersResponse)
				if err != nil {
					logger.Error("failed-to-decode-response", err)
					return err
				}

				users = listUsersResponse.Resources
				for _, u := range users {
					roleAssignments <- RoleAssignment{
						RoleName:     roleRequest.Role,
						ResourceGUID: space.Metadata.GUID,
						UserGUID:     u.Metadata.GUID,
					}
				}

				return nil
			})
			if err != nil {
				return err
			}
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
