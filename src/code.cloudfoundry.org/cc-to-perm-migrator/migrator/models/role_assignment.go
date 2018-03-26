package models

type RoleAssignment struct {
	ResourceGUID string
	UserGUID     string
	Roles        []string
}
