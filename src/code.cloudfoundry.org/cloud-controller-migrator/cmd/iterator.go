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

func IterateOverCloudControllerEntities(ctx context.Context, logger lager.Logger, w io.Writer, client *http.Client, url string) error {
	logger = logger.Session("iterate-over-cloud-controller-entities").WithData(lager.Data{
		"url": url,
	})

	rg := cloudcontroller.NewRequestGenerator(url)

	routerLogger := logger.WithData(lager.Data{
		"routes": rg.Routes,
	})

	var (
		route string
		req   *http.Request
		res   *http.Response
		err   error
	)

	var (
		routeLogger lager.Logger
	)

	// /v2/info - Equivalent to Ping
	route = cloudcontroller.Info
	routeLogger = routerLogger.WithData(lager.Data{
		"route": route,
	})
	req, err = rg.CreateRequest(route, nil, nil)
	if err != nil {
		routeLogger.Error(messages.FailedToCreateRequest, err)
		return err
	}

	res, err = client.Do(req)
	if err != nil {
		routeLogger.Error(messages.FailedToPerformRequest, err)

		return err
	}

	if res.StatusCode >= 400 {
		err = fmt.Errorf("HTTP bad response: %d", res.StatusCode)
		routeLogger.Error("failed-to-ping-cloudcontroller", err)
		return err
	}

	// List Organizations
	var organizations []OrganizationResource

	route = cloudcontroller.ListOrganizations
	routeLogger = routerLogger.WithData(lager.Data{
		"route": route,
	})
	req, err = rg.CreateRequest(route, nil, nil)
	if err != nil {
		routeLogger.Error(messages.FailedToCreateRequest, err)
		return err
	}

	res, err = client.Do(req)
	if err != nil {
		routeLogger.Error(messages.FailedToPerformRequest, err)
		return err
	}

	defer res.Body.Close()

	var listOrganizationsResponse ListOrganizationsResponse
	err = json.NewDecoder(res.Body).Decode(&listOrganizationsResponse)
	if err != nil {
		routeLogger.Error("failed-to-decode-response", err)
	}

	organizations = append(organizations, listOrganizationsResponse.Resources...)

	for _, organization := range organizations {
		fmt.Fprintf(w, "%s: %s\n", organization.Metadata.GUID, organization.Entity.Name)
	}

	return nil
}

type ListOrganizationsResponse struct {
	Resources []OrganizationResource `json:"resources"`
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
