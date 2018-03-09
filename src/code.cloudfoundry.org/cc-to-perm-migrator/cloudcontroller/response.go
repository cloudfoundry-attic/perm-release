package cloudcontroller

type PaginatedResponse struct {
	NextURL     *string `json:"next_url"`
}

type ListOrganizationsResponse struct {
	PaginatedResponse
	Resources []OrganizationResource `json:"resources"`
}

type OrganizationResource struct {
	MetadataResource
}

type ListSpacesResponse struct {
	PaginatedResponse
	Resources []SpaceResource `json:"resources"`
}

type SpaceResource struct {
	MetadataResource
}

type MetadataResource struct {
	Metadata struct {
		GUID string `json:"guid"`
	} `json:"metadata"`
}

type ListOrganizationRolesResponse struct {
	PaginatedResponse
	Resources []OrgUserResource `json:"resources"`
}

type ListSpaceRolesResponse struct {
	PaginatedResponse
	Resources []SpaceUserResource `json:"resources"`
}

type OrgUserResource struct {
	MetadataResource
	Entity struct {
		Roles []string `json:"organization_roles"`
	}
}
type SpaceUserResource struct {
	MetadataResource
	Entity struct {
		Roles []string `json:"space_roles"`
	}
}

