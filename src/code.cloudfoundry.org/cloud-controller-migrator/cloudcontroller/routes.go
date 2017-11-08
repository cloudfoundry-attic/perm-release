package cloudcontroller

import (
	"net/http"

	"github.com/tedsuo/rata"
)

var Routes = rata.Routes{
	{Name: Info, Path: "/v2/info", Method: rata.GET},

	{Name: ListOrganizations, Path: "/v2/organizations", Method: rata.GET},
	{Name: ListOrganizationSpaces, Path: "/v2/organizations/:guid/spaces", Method: rata.GET},
	{Name: ListOrganizationAuditors, Path: "/v2/organizations/:guid/auditors", Method: rata.GET},
	{Name: ListOrganizationBillingManagers, Path: "/v2/organizations/:guid/billing_managers", Method: rata.GET},
	{Name: ListOrganizationManagers, Path: "/v2/organizations/:guid/managers", Method: rata.GET},
	{Name: ListOrganizationUsers, Path: "/v2/organizations/:guid/users", Method: rata.GET},

	{Name: ListSpaceAuditors, Path: "/v2/spaces/:guid/auditors", Method: rata.GET},
	{Name: ListSpaceDevelopers, Path: "/v2/spaces/:guid/developers", Method: rata.GET},
	{Name: ListSpaceManagers, Path: "/v2/spaces/:guid/managers", Method: rata.GET},
}

const (
	Info = "info"

	ListOrganizations = "list_organizations"

	ListOrganizationSpaces = "list_organization_spaces"

	ListOrganizationAuditors        = "list_organization_auditors"
	ListOrganizationBillingManagers = "list_organization_billing_managers"
	ListOrganizationManagers        = "list_organization_managers"
	ListOrganizationUsers           = "list_organization_users"

	ListSpaceAuditors   = "list_space_auditors"
	ListSpaceDevelopers = "list_space_developers"
	ListSpaceManagers   = "list_space_managers"
)

type RequestGenerator struct {
	*rata.RequestGenerator

	Routes rata.Routes
}

func NewRequestGenerator(host string) *RequestGenerator {
	rg := rata.NewRequestGenerator(host, Routes)

	header := http.Header{}
	header.Add("Accept", "application/json")

	rg.Header = header

	return &RequestGenerator{
		RequestGenerator: rg,
		Routes:           Routes,
	}
}
