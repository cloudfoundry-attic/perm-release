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

type tokenJSON struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"` // at least PayPal returns string, while most return number
	Expires      string `json:"expires"`    // broken Facebook spelling of expires_in
}

func appendHandler(server *ghttp.Server, method, path string, response interface{}) {
	server.AppendHandlers(
		ghttp.CombineHandlers(
			ghttp.VerifyRequest(method, path, ""),
			ghttp.RespondWithJSONEncoded(http.StatusOK, response),
		),
	)
}

var _ = Describe("CCToPermMigrator", func() {
	var (
		server         *ghttp.Server
		nextSpacesPage = "/v2/spaces/page2"

		listSpacesResponse1Part1 capimodels.ListSpacesResponse
		listSpacesResponse1Part2 capimodels.ListSpacesResponse
		listSpacesResponse2      capimodels.ListSpacesResponse
		listSpaceRolesNextPage   = "/v2/spaces/next"

		listSpaceRolesResponse1Part1 capimodels.ListSpaceRolesResponse
		listSpaceRolesResponse1Part2 capimodels.ListSpaceRolesResponse
		listSpaceRolesResponse2      capimodels.ListSpaceRolesResponse
		listSpaceRolesResponse3      capimodels.ListSpaceRolesResponse
		nextOrgRolesPage             = "/v2/organizations/guid/next"

		orgResponsePage1 capimodels.ListOrgsResponse
		orgResponsePage2 capimodels.ListOrgsResponse
		nextOrgsPage     = "/v2/organizations/page2"

		listOrgRolesResponse1Part1 capimodels.ListOrgRolesResponse
		listOrgRolesResponse1Part2 capimodels.ListOrgRolesResponse
		listOrgRolesResponse2      capimodels.ListOrgRolesResponse
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		listSpacesResponse1Part1 = capimodels.ListSpacesResponse{
			PaginatedResponse: capimodels.PaginatedResponse{
				NextURL: &nextSpacesPage,
			},
			Resources: []capimodels.SpaceResource{
				{
					Metadata: capimodels.MetadataResource{
						GUID: "space-guid-1",
					},
				},
			},
		}
		listSpacesResponse1Part2 = capimodels.ListSpacesResponse{
			Resources: []capimodels.SpaceResource{
				{
					Metadata: capimodels.MetadataResource{
						GUID: "space-guid-2",
					},
				},
			},
		}
		listSpacesResponse2 = capimodels.ListSpacesResponse{
			Resources: []capimodels.SpaceResource{
				{
					Metadata: capimodels.MetadataResource{
						GUID: "space-guid-3",
					},
				},
			},
		}

		listSpaceRolesResponse1Part1 = capimodels.ListSpaceRolesResponse{
			PaginatedResponse: capimodels.PaginatedResponse{
				NextURL: &listSpaceRolesNextPage,
			},
			Resources: []capimodels.SpaceUserResource{
				{
					Metadata: capimodels.MetadataResource{
						GUID: "test-user-1",
					},
					Entity: capimodels.SpaceUserResourceEntity{
						Roles: []string{"space_developer"},
					},
				},
			},
		}
		listSpaceRolesResponse1Part2 = capimodels.ListSpaceRolesResponse{
			Resources: []capimodels.SpaceUserResource{
				{
					Metadata: capimodels.MetadataResource{
						GUID: "test-user-3",
					},
					Entity: capimodels.SpaceUserResourceEntity{
						Roles: []string{"space_manager"},
					},
				},
			},
		}
		listSpaceRolesResponse2 = capimodels.ListSpaceRolesResponse{
			Resources: []capimodels.SpaceUserResource{
				{
					Metadata: capimodels.MetadataResource{
						GUID: "test-user-3",
					},
					Entity: capimodels.SpaceUserResourceEntity{
						Roles: []string{"space_manager"},
					},
				},
			},
		}

		listSpaceRolesResponse3 = capimodels.ListSpaceRolesResponse{
			Resources: []capimodels.SpaceUserResource{
				{
					Metadata: capimodels.MetadataResource{
						GUID: "test-user-3",
					},
					Entity: capimodels.SpaceUserResourceEntity{
						Roles: []string{"space_auditor"},
					},
				},
			},
		}

		orgResponsePage1 = capimodels.ListOrgsResponse{
			PaginatedResponse: capimodels.PaginatedResponse{
				NextURL: &nextOrgsPage,
			},
			Resources: []capimodels.OrgResource{{
				Metadata: capimodels.MetadataResource{
					GUID: "org-guid-1",
				},
			}},
		}

		orgResponsePage2 = capimodels.ListOrgsResponse{
			Resources: []capimodels.OrgResource{{
				Metadata: capimodels.MetadataResource{
					GUID: "org-guid-2",
				},
			}},
		}

		listOrgRolesResponse1Part1 = capimodels.ListOrgRolesResponse{
			PaginatedResponse: capimodels.PaginatedResponse{
				NextURL: &nextOrgRolesPage,
			},
			Resources: []capimodels.OrgUserResource{
				{
					Metadata: capimodels.MetadataResource{
						GUID: "test-user-1",
					},
					Entity: capimodels.OrgUserResourceEntity{
						Roles: []string{"org_developer"},
					},
				},
			},
		}
		listOrgRolesResponse1Part2 = capimodels.ListOrgRolesResponse{
			Resources: []capimodels.OrgUserResource{
				{
					Metadata: capimodels.MetadataResource{
						GUID: "test-user-1",
					},
					Entity: capimodels.OrgUserResourceEntity{
						Roles: []string{"org_developer"},
					},
				},
			},
		}
		listOrgRolesResponse2 = capimodels.ListOrgRolesResponse{
			Resources: []capimodels.OrgUserResource{
				{
					Metadata: capimodels.MetadataResource{
						GUID: "test-user-1",
					},
					Entity: capimodels.OrgUserResourceEntity{
						Roles: []string{"org_developer"},
					},
				},
			},
		}
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("#Main", func() {
		var configFilePath string
		var tmpDir string
		var err error

		BeforeEach(func() {
			appendHandler(server, "POST", "/oauth/token", tokenJSON{
				AccessToken:  "cool",
				TokenType:    "whatever",
				RefreshToken: "something",
				ExpiresIn:    "1234",
				Expires:      "1234",
			})
			appendHandler(server, "GET", "/v2/organizations", orgResponsePage1)
			appendHandler(server, "GET", nextOrgsPage, orgResponsePage2)
			appendHandler(server, "GET", "/v2/organizations/org-guid-1/user_roles", listOrgRolesResponse1Part1)
			appendHandler(server, "GET", nextOrgRolesPage, listOrgRolesResponse1Part2)
			appendHandler(server, "GET", "/v2/organizations/org-guid-1/spaces", listSpacesResponse1Part1)
			appendHandler(server, "GET", nextSpacesPage, listSpacesResponse1Part2)
			appendHandler(server, "GET", "/v2/spaces/space-guid-1/user_roles", listSpaceRolesResponse1Part1)
			appendHandler(server, "GET", listSpaceRolesNextPage, listSpaceRolesResponse1Part2)

			appendHandler(server, "GET", "/v2/spaces/space-guid-2/user_roles", listSpaceRolesResponse2)
			appendHandler(server, "GET", "/v2/organizations/org-guid-2/user_roles", listOrgRolesResponse2)
			appendHandler(server, "GET", "/v2/organizations/org-guid-2/spaces", listSpacesResponse2)
			appendHandler(server, "GET", "/v2/spaces/space-guid-3/user_roles", listSpaceRolesResponse3)

			tmpDir, err = ioutil.TempDir("", "ccmtest")
			Expect(err).NotTo(HaveOccurred())
			configFilePath = path.Join(tmpDir, "config.yml")

			contents := fmt.Sprintf(`log_level: info

uaa:
  url: %s
  ca_cert_path: /var/vcap/jobs/cc-to-perm-migrator/config/certs/uaa-ca.crt

cloud_controller:
  url: %s
  client_id: perm-migrator
  client_secret: secret
  client_scopes: ["cloud_controller.admin_read_only"]
`, server.URL(), server.URL())
			err = ioutil.WriteFile(configFilePath, []byte(contents), 0600)
		})
		It("exits with 1 when no flags are passed", func() {
			session := RunCommand("--config-file-path", configFilePath)
			Eventually(session, 1).Should(gexec.Exit(0))
		})
	})
})
