package main_test

import (
	"net/http"

	"fmt"

	"io/ioutil"
	"path"

	"code.cloudfoundry.org/cc-to-perm-migrator/capi/capimodels"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

const configTemplate = `log_level: info

uaa:
  url: %s
  ca_cert_path: /var/vcap/jobs/cc-to-perm-migrator/config/certs/uaa-ca.crt

cloud_controller:
  url: %s
  client_id: perm-migrator
  client_secret: secret
  client_scopes: ["cloud_controller.admin_read_only"]
`

var _ = Describe("CCToPermMigrator", func() {
	var (
		server *ghttp.Server

		orgsPage1 capimodels.ListOrgsResponse
		orgsPage2 capimodels.ListOrgsResponse

		orgRoles1Page1   capimodels.ListOrgRolesResponse
		orgRoles1Page2   capimodels.ListOrgRolesResponse
		spaces1Page1     capimodels.ListSpacesResponse
		spaces1Page2     capimodels.ListSpacesResponse
		spaceRoles1Page1 capimodels.ListSpaceRolesResponse
		spaceRoles1Page2 capimodels.ListSpaceRolesResponse
		spaceRoles2      capimodels.ListSpaceRolesResponse

		orgRoles2   capimodels.ListOrgRolesResponse
		spaces2     capimodels.ListSpacesResponse
		spaceRoles3 capimodels.ListSpaceRolesResponse

		orgsNextURL       = "/v2/organizations/page2"
		orgRolesNextURL   = "/v2/organizations/guid/next"
		spacesNextURL     = "/v2/spaces/page2"
		spaceRolesNextURL = "/v2/spaces/next"
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		orgsPage1 = createOrgResponse([]string{"org-guid-1"}, orgsNextURL)
		orgsPage2 = createOrgResponse([]string{"org-guid-2"}, "")
		orgRoles1Page1 = createOrgRolesResponse([]string{"test-user-1"}, orgRolesNextURL, []string{"org_developer"})
		orgRoles1Page2 = createOrgRolesResponse([]string{"test-user-2"}, "", []string{"org_manager"})
		spaces1Page1 = createSpacesResponse([]string{"space-guid-1"}, spacesNextURL)
		spaces1Page2 = createSpacesResponse([]string{"space-guid-2"}, "")
		spaceRoles1Page1 = createSpaceRolesResponse([]string{"test-user-1"}, spaceRolesNextURL, []string{"space_developer"})
		spaceRoles1Page2 = createSpaceRolesResponse([]string{"test-user-2"}, "", []string{"space_manager"})
		spaceRoles2 = createSpaceRolesResponse([]string{"test-user-3"}, "", []string{"space_developer"})

		orgRoles2 = createOrgRolesResponse([]string{"test-user-1"}, "", []string{"billing_manager"})
		spaces2 = createSpacesResponse([]string{"space-guid-3"}, "")
		spaceRoles3 = createSpaceRolesResponse([]string{"test-user-1"}, "", []string{"space_auditor"})
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("#Main", func() {
		var configFilePath string
		var tmpDir string
		var err error

		BeforeEach(func() {
			//These handlers are appended in the order in which they are called.
			//If adding more handlers, make sure they are placed correctly in the set of calls.
			appendHandler(server, "POST", "/oauth/token", tokenJSON{
				AccessToken:  "cool",
				TokenType:    "whatever",
				RefreshToken: "something",
				ExpiresIn:    "1234",
			})
			appendHandler(server, "GET", "/v2/organizations", orgsPage1)
			appendHandler(server, "GET", orgsNextURL, orgsPage2)
			appendHandler(server, "GET", "/v2/organizations/org-guid-1/user_roles", orgRoles1Page1)
			appendHandler(server, "GET", orgRolesNextURL, orgRoles1Page2)
			appendHandler(server, "GET", "/v2/organizations/org-guid-1/spaces", spaces1Page1)
			appendHandler(server, "GET", spacesNextURL, spaces1Page2)
			appendHandler(server, "GET", "/v2/spaces/space-guid-1/user_roles", spaceRoles1Page1)
			appendHandler(server, "GET", spaceRolesNextURL, spaceRoles1Page2)

			appendHandler(server, "GET", "/v2/spaces/space-guid-2/user_roles", spaceRoles2)
			appendHandler(server, "GET", "/v2/organizations/org-guid-2/user_roles", orgRoles2)
			appendHandler(server, "GET", "/v2/organizations/org-guid-2/spaces", spaces2)
			appendHandler(server, "GET", "/v2/spaces/space-guid-3/user_roles", spaceRoles3)

			tmpDir, err = ioutil.TempDir("", "ccmtest")
			Expect(err).NotTo(HaveOccurred())
			configFilePath = path.Join(tmpDir, "config.yml")

			contents := fmt.Sprintf(configTemplate, server.URL(), server.URL())
			err = ioutil.WriteFile(configFilePath, []byte(contents), 0600)
		})

		It("exits with 1 when no flags are passed", func() {
			session := RunCommand("--config-file-path", configFilePath)
			Eventually(session, 1).Should(gexec.Exit(0))
		})
	})
})

func appendHandler(server *ghttp.Server, method, path string, response interface{}) {
	server.AppendHandlers(
		ghttp.CombineHandlers(
			ghttp.VerifyRequest(method, path, ""),
			ghttp.RespondWithJSONEncoded(http.StatusOK, response),
		),
	)
}

func createOrgResponse(GUIDs []string, nextPageURL string) capimodels.ListOrgsResponse {
	response := capimodels.ListOrgsResponse{}
	for _, guid := range GUIDs {
		orgResource := capimodels.OrgResource{
			Metadata: capimodels.MetadataResource{GUID: guid},
		}
		response.Resources = append(response.Resources, orgResource)
	}
	if nextPageURL != "" {
		response.PaginatedResponse = capimodels.PaginatedResponse{
			NextURL: &nextPageURL,
		}
	}
	return response

}

func createOrgRolesResponse(GUIDs []string, nextPageURL string, roles []string) capimodels.ListOrgRolesResponse {
	//Note: All resources are populated with the same roles
	response := capimodels.ListOrgRolesResponse{}
	for _, guid := range GUIDs {
		orgUserResource := capimodels.OrgUserResource{
			Metadata: capimodels.MetadataResource{GUID: guid},
			Entity:   capimodels.OrgUserResourceEntity{Roles: roles},
		}
		response.Resources = append(response.Resources, orgUserResource)
	}
	if nextPageURL != "" {
		response.PaginatedResponse = capimodels.PaginatedResponse{
			NextURL: &nextPageURL,
		}
	}
	return response
}

func createSpacesResponse(GUIDs []string, nextPageURL string) capimodels.ListSpacesResponse {
	response := capimodels.ListSpacesResponse{}
	for _, guid := range GUIDs {
		spaceResource := capimodels.SpaceResource{
			Metadata: capimodels.MetadataResource{GUID: guid},
		}
		response.Resources = append(response.Resources, spaceResource)
	}
	if nextPageURL != "" {
		response.PaginatedResponse = capimodels.PaginatedResponse{
			NextURL: &nextPageURL,
		}
	}
	return response
}

func createSpaceRolesResponse(GUIDs []string, nextPageURL string, roles []string) capimodels.ListSpaceRolesResponse {
	//Note: All resources are populated with the same roles
	response := capimodels.ListSpaceRolesResponse{}
	for _, guid := range GUIDs {
		spaceUserResource := capimodels.SpaceUserResource{
			Metadata: capimodels.MetadataResource{GUID: guid},
			Entity:   capimodels.SpaceUserResourceEntity{Roles: roles},
		}
		response.Resources = append(response.Resources, spaceUserResource)
	}
	if nextPageURL != "" {
		response.PaginatedResponse = capimodels.PaginatedResponse{
			NextURL: &nextPageURL,
		}
	}
	return response
}

type tokenJSON struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
}
