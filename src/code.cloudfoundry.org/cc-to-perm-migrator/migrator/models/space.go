package models

type Space struct {
	GUID        string
	OrgGUID     string
	Assignments []RoleAssignment
}
