package capi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"time"

	"code.cloudfoundry.org/cc-to-perm-migrator/capi/capimodels"
	"code.cloudfoundry.org/cc-to-perm-migrator/messages"
	"code.cloudfoundry.org/cc-to-perm-migrator/migrator"
	"code.cloudfoundry.org/lager"
)

type Client struct {
	host       string
	httpClient *http.Client
}

func NewClient(host string) *Client {
	return &Client{
		host:       host,
		httpClient: http.DefaultClient,
	}
}

func (c *Client) GetOrgRoleAssignments(logger lager.Logger, orgGUID string) ([]migrator.RoleAssignment, error) {
	route := fmt.Sprintf("/v2/organizations/%s/user_roles", orgGUID)

	var orgRoles []migrator.RoleAssignment

	err := c.makePaginatedGetRequest(logger, route, func(logger lager.Logger, r io.Reader) error {
		var listOrgRolesResponse capimodels.ListOrgRolesResponse
		err := json.NewDecoder(r).Decode(&listOrgRolesResponse)
		if err != nil {
			logger.Error("failed-to-decode-response", err)
			return err
		}
		for _, role := range listOrgRolesResponse.Resources {
			orgRoles = append(orgRoles, migrator.RoleAssignment{
				UserGUID:     role.Metadata.GUID,
				ResourceGUID: orgGUID,
				Roles:        role.Entity.Roles,
			})
		}

		return nil
	})

	if err != nil {
		return orgRoles, errors.New("failed-to-fetch-organization-user-roles")
	}

	return orgRoles, nil
}

func (c *Client) GetSpaceRoleAssignments(logger lager.Logger, spaceGUID string) ([]migrator.RoleAssignment, error) {
	route := fmt.Sprintf("/v2/spaces/%s/user_roles", spaceGUID)

	var spaceRoles []migrator.RoleAssignment

	err := c.makePaginatedGetRequest(logger, route, func(logger lager.Logger, r io.Reader) error {
		var listSpaceRolesResponse capimodels.ListSpaceRolesResponse
		err := json.NewDecoder(r).Decode(&listSpaceRolesResponse)
		if err != nil {
			logger.Error("failed-to-decode-response", err)
			return err
		}
		for _, role := range listSpaceRolesResponse.Resources {
			spaceRoles = append(spaceRoles, migrator.RoleAssignment{
				UserGUID:     role.Metadata.GUID,
				ResourceGUID: spaceGUID,
				Roles:        role.Entity.Roles,
			})
		}

		return nil
	})

	if err != nil {
		return spaceRoles, errors.New("failed-to-fetch-space-user-roles")
	}
	return spaceRoles, nil
}

func (c *Client) makePaginatedGetRequest(logger lager.Logger, route string, bodyCallback func(lager.Logger, io.Reader) error) error {
	rg := NewRequestGenerator(c.host)

	var (
		paginatedResponse capimodels.PaginatedResponse

		routeLogger lager.Logger
	)

	for {
		nextURL, err := func() (*string, error) {
			routeLogger = logger.WithData(lager.Data{
				"route": route,
			})

			ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
			defer cancelFunc()

			res, err := makeAPIRequest(ctx, routeLogger.Session("make-api-request"), c.httpClient, rg, route)
			if err != nil {
				return nil, err
			}
			defer res.Body.Close()

			var body []byte
			buf := bytes.NewBuffer(body)
			r := io.TeeReader(res.Body, buf)

			err = json.NewDecoder(r).Decode(&paginatedResponse)
			if err != nil {
				routeLogger.Error("failed-to-decode-response", err)
				return nil, err
			}

			err = bodyCallback(routeLogger, buf)
			if err != nil {
				return nil, err
			}

			if paginatedResponse.NextURL == nil {
				return nil, nil
			}

			return paginatedResponse.NextURL, nil
		}()

		if err != nil {
			return err
		}

		if nextURL == nil {
			break
		}

		route = *nextURL
	}

	return nil
}

func makeAPIRequest(ctx context.Context, logger lager.Logger, client *http.Client, rg *RequestGenerator, route string) (*http.Response, error) {
	req, err := rg.NewGetRequest(logger.Session("new-get-request"), route)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

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
