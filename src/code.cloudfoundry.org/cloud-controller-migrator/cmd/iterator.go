package cmd

import (
	"context"
	"io"
	"net/http"

	"fmt"

	"encoding/json"

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
		res   *http.Response
		err   error
	)

	var (
		routeLogger lager.Logger
	)

	// /v2/info - Equivalent to Ping
	route = "/v2/info"
	routeLogger = logger.WithData(lager.Data{
		"route": route,
	})
	res, err = makeAPIRequest(routeLogger, client, rg, route)
	if err != nil {
		return err
	}

	// List Organizations
	route = "/v2/organizations"

	var organizations []cloudcontroller.OrganizationResource

	for {
		routeLogger = logger.WithData(lager.Data{
			"route": route,
		})

		res, err = makeAPIRequest(routeLogger, client, rg, route)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		var listOrganizationsResponse cloudcontroller.ListOrganizationsResponse
		err = json.NewDecoder(res.Body).Decode(&listOrganizationsResponse)
		if err != nil {
			routeLogger.Error("failed-to-decode-response", err)
		}

		organizations = append(organizations, listOrganizationsResponse.Resources...)
		if listOrganizationsResponse.NextURL == nil {
			break
		} else {
			route = *listOrganizationsResponse.NextURL
		}
	}

	var spaces []cloudcontroller.SpaceResource

	for _, organization := range organizations {
		route = organization.Entity.SpacesURL

		for {
			routeLogger = logger.WithData(lager.Data{
				"route": route,
			})

			res, err = makeAPIRequest(routeLogger, client, rg, route)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			var listOrganizationSpacesResponse cloudcontroller.ListOrganizationSpacesResponse
			err = json.NewDecoder(res.Body).Decode(&listOrganizationSpacesResponse)
			if err != nil {
				routeLogger.Error("failed-to-decode-response", err)
			}

			spaces = append(spaces, listOrganizationSpacesResponse.Resources...)
			if listOrganizationSpacesResponse.NextURL == nil {
				break
			} else {
				route = *listOrganizationSpacesResponse.NextURL
			}
		}
	}

	fmt.Fprintf(w, "Organizations: %d\n", len(organizations))
	fmt.Fprintf(w, "Average spaces per organization: %f\n", float32(len(spaces))/float32(len(organizations)))

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
