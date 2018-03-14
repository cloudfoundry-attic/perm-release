package capimodels

type PaginatedResponse struct {
	NextURL *string `json:"next_url"`
}

type ListOrgsResponse struct {
	PaginatedResponse
	Resources []OrgResource `json:"resources"`
}

type OrgResource struct {
	Metadata MetadataResource `json:"metadata"`
}

type ListSpacesResponse struct {
	PaginatedResponse
	Resources []SpaceResource `json:"resources"`
}

type SpaceResource struct {
	Metadata MetadataResource `json:"metadata"`
}

type MetadataResource struct {
	GUID string `json:"guid"`
}

type ListOrgRolesResponse struct {
	PaginatedResponse
	Resources []OrgUserResource `json:"resources"`
}

type ListSpaceRolesResponse struct {
	PaginatedResponse
	Resources []SpaceUserResource `json:"resources"`
}

type OrgUserResource struct {
	Metadata MetadataResource      `json:"metadata"`
	Entity   OrgUserResourceEntity `json:"entity"`
}

type SpaceUserResource struct {
	Metadata MetadataResource        `json:"metadata"`
	Entity   SpaceUserResourceEntity `json:"entity"`
}

type SpaceUserResourceEntity struct {
	Roles []string `json:"space_roles"`
}

type OrgUserResourceEntity struct {
	Roles []string `json:"organization_roles"`
}
