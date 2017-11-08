package cmd

import (
	"context"
	"io"

	"fmt"

	"encoding/json"

	"code.cloudfoundry.org/cloud-controller-migrator/cloudcontroller"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter . CloudControllerAPIClient

type CloudControllerAPIClient interface {
	MakePaginatedGetRequest(ctx context.Context, logger lager.Logger, route string, bodyCallback func(context.Context, lager.Logger, io.Reader) error) error
}

func IterateOverCloudControllerEntities(ctx context.Context, logger lager.Logger, w io.Writer, c CloudControllerAPIClient) error {
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
	var organizationRoleAssignments []RoleAssignment

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

		var (
			roleAssignment RoleAssignment
		)

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
					roleAssignment = RoleAssignment{
						RoleName:     roleRequest.Role,
						ResourceGUID: organization.Metadata.GUID,
						UserGUID:     u.Metadata.GUID,
					}
					organizationRoleAssignments = append(organizationRoleAssignments, roleAssignment)
				}

				return nil
			})
			if err != nil {
				return err
			}
		}

	}

	var spaceRoleAssignments []RoleAssignment

	for _, space := range spaces {
		var (
			roleAssignment RoleAssignment
		)

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
					roleAssignment = RoleAssignment{
						RoleName:     roleRequest.Role,
						ResourceGUID: space.Metadata.GUID,
						UserGUID:     u.Metadata.GUID,
					}
					spaceRoleAssignments = append(spaceRoleAssignments, roleAssignment)
				}

				return nil
			})
			if err != nil {
				return err
			}
		}

	}

	fmt.Fprintf(w, "\nReport\n==========================================\n")
	fmt.Fprintf(w, "Organizations: %d\n", len(organizations))
	fmt.Fprintf(w, "Average spaces per organization: %f\n", float32(len(spaces))/float32(len(organizations)))
	fmt.Fprintf(w, "Average role assignments per organization: %f\n", float32(len(organizationRoleAssignments))/float32(len(organizations)))
	fmt.Fprintf(w, "Average role assignments per space: %f\n", float32(len(spaceRoleAssignments))/float32(len(spaces)))

	return nil
}

type RoleAssignment struct {
	RoleName     string
	ResourceGUID string
	UserGUID     string
}
