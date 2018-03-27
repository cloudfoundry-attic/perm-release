package models

type Organization struct {
	GUID        string
	Assignments []RoleAssignment
}
