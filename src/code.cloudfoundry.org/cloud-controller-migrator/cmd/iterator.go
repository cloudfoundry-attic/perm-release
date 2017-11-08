package cmd

import (
	"context"
	"io"
	"net/http"

	"fmt"

	"encoding/json"

	"bytes"

	"code.cloudfoundry.org/cloud-controller-migrator/cloudcontroller"
	"code.cloudfoundry.org/cloud-controller-migrator/messages"
	"code.cloudfoundry.org/lager"
)

func IterateOverCloudControllerEntities(ctx context.Context, logger lager.Logger, w io.Writer, client *http.Client, host string) error {
	logger = logger.Session("iterate-over-cloud-controller-entities").WithData(lager.Data{
		"host": host,
	})

	rg := cloudcontroller.NewRequestGenerator(host)

	var (
		route string
		err   error
	)

	// List Organizations
	route = "/v2/organizations"

	var organizations []cloudcontroller.OrganizationResource

	err = makePaginatedAPIRequest(logger, client, rg, route, func(logger lager.Logger, r io.Reader) error {
		var listOrganizationsResponse cloudcontroller.ListOrganizationsResponse
		err = json.NewDecoder(r).Decode(&listOrganizationsResponse)
		if err != nil {
			logger.Error("failed-to-decode-response", err)
			return err
		}

		organizations = append(organizations, listOrganizationsResponse.Resources...)
		return nil
	})

	var spaces []cloudcontroller.SpaceResource
	var organizationRoleAssignments []RoleAssignment

	for _, organization := range organizations {
		route = organization.Entity.SpacesURL

		err = makePaginatedAPIRequest(logger, client, rg, route, func(logger lager.Logger, r io.Reader) error {
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

		type RoleRequest struct {
			Route string
			Role  string
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

			err = makePaginatedAPIRequest(logger, client, rg, route, func(logger lager.Logger, r io.Reader) error {
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
		}
	}

	fmt.Fprintf(w, "Organizations: %d\n", len(organizations))
	fmt.Fprintf(w, "Average spaces per organization: %f\n", float32(len(spaces))/float32(len(organizations)))
	fmt.Fprintf(w, "Average role assignments per organization: %f\n", float32(len(organizationRoleAssignments))/float32(len(organizations)))

	return nil
}

func makePaginatedAPIRequest(logger lager.Logger, client *http.Client, rg *cloudcontroller.RequestGenerator, route string, bodyCallback func(lager.Logger, io.Reader) error) error {
	var (
		res *http.Response
		err error

		paginatedResponse cloudcontroller.PaginatedResponse

		routeLogger lager.Logger
	)

	for {
		routeLogger = logger.WithData(lager.Data{
			"route": route,
		})

		res, err = makeAPIRequest(routeLogger.Session("make-api-request"), client, rg, route)
		if err != nil {
			return err
		}

		var body []byte
		buf := bytes.NewBuffer(body)
		r := io.TeeReader(res.Body, buf)

		defer res.Body.Close()

		err = json.NewDecoder(r).Decode(&paginatedResponse)
		if err != nil {
			return err
		}

		err = bodyCallback(routeLogger, buf)
		if err != nil {
			return err
		}

		if paginatedResponse.NextURL == nil {
			break
		} else {
			route = *paginatedResponse.NextURL
		}
	}

	return nil
}

func makeAPIRequest(logger lager.Logger, client *http.Client, rg *cloudcontroller.RequestGenerator, route string) (*http.Response, error) {
	req, err := rg.NewGetRequest(logger.Session("new-get-request"), route)
	if err != nil {
		return nil, err
	}

	logger.Debug("making-request")
	res, err := client.Do(req)
	if err != nil {
		logger.Error(messages.FailedToPerformRequest, err)
		return nil, err
	}

	if res.StatusCode >= 400 {
		err = fmt.Errorf("HTTP bad response: %d", res.StatusCode)
		logger.Error("failed-to-ping-cloudcontroller", err)
		return nil, err
	}

	return res, nil
}

type RoleAssignment struct {
	RoleName     string
	ResourceGUID string
	UserGUID     string
}
