package populator

import (
	"context"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm-go"
)

const (
	orgAuditor        = "auditor"
	orgBillingManager = "billing_manager"
	orgManager        = "manager"
	orgUser           = "user"

	spaceAuditor   = "auditor"
	spaceDeveloper = "developer"
	spaceManager   = "manager"
)

type Populator struct {
	client protos.RoleServiceClient
}

func NewPopulator(client protos.RoleServiceClient) *Populator {
	return &Populator{
		client: client,
	}
}

func (p *Populator) PopulateOrganization(logger lager.Logger, org models.Organization, namespace string) []error {
	var errs []error

	roles := []string{orgAuditor, orgBillingManager, orgManager, orgUser}

	for _, role := range roles {
		req := &protos.CreateRoleRequest{
			Name: makeOrgRoleName(role, org.GUID),
			Permissions: []*protos.Permission{
				{
					Name:            fmt.Sprintf("org.%s", role),
					ResourcePattern: org.GUID,
				},
			},
		}

		if _, err := p.client.CreateRole(context.Background(), req); err != nil {
			errs = append(errs, err)
		}
	}

	for _, assignment := range org.Assignments {
		actor := &protos.Actor{
			ID:     assignment.UserGUID,
			Issuer: namespace,
		}

		for _, role := range assignment.Roles {
			req := &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: makeOrgRoleName(role, org.GUID),
			}

			if _, err := p.client.AssignRole(context.Background(), req); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errs
}

func (p *Populator) PopulateSpace(logger lager.Logger, space models.Space, namespace string) []error {
	var errs []error

	roles := []string{spaceAuditor, spaceDeveloper, spaceManager}

	for _, role := range roles {
		req := &protos.CreateRoleRequest{
			Name: makeSpaceRoleName(role, space.GUID),
			Permissions: []*protos.Permission{
				{
					Name:            fmt.Sprintf("space.%s", role),
					ResourcePattern: space.GUID,
				},
			},
		}

		if _, err := p.client.CreateRole(context.Background(), req); err != nil {
			errs = append(errs, err)
		}
	}

	for _, assignment := range space.Assignments {
		actor := &protos.Actor{
			ID:     assignment.UserGUID,
			Issuer: namespace,
		}

		for _, role := range assignment.Roles {
			req := &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: makeSpaceRoleName(role, space.GUID),
			}

			if _, err := p.client.AssignRole(context.Background(), req); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errs
}

func makeOrgRoleName(role, id string) string {
	role = strings.Replace(role, "org-", "", -1)
	role = strings.Replace(role, "org_", "", -1)

	return fmt.Sprintf("org-%s-%s", role, id)
}

func makeSpaceRoleName(role, id string) string {
	role = strings.Replace(role, "space-", "", -1)
	role = strings.Replace(role, "space_", "", -1)

	return fmt.Sprintf("space-%s-%s", role, id)
}
