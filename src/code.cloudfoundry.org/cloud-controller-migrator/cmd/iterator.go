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

	var organizations []OrganizationResource

	for {
		routeLogger = logger.WithData(lager.Data{
			"route": route,
		})

		res, err = makeAPIRequest(routeLogger, client, rg, route)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		var listOrganizationsResponse ListOrganizationsResponse
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

	for _, organization := range organizations {
		fmt.Fprintf(w, "%s: %s\n", organization.Metadata.GUID, organization.Entity.Name)
	}

	return nil
}

type ListOrganizationsResponse struct {
	NextURL     *string                `json:"next_url"`
	PreviousURL *string                `json:"prev_url"`
	Resources   []OrganizationResource `json:"resources"`
}

type OrganizationResource struct {
	Metadata struct {
		GUID string `json:"guid"`
		URL  string `json:"url"`
	} `json:"metadata"`
	Entity struct {
		Name               string `json:"name"`
		SpacesURL          string `json:"spaces_url"`
		UsersURL           string `json:"users_url"`
		ManagersURL        string `json:"managers_url"`
		BillingManagersURL string `json:"billing_managers_url"`
		AuditorsURL        string `json:"auditors_url"`
	} `json:"entity"`
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
