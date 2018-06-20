package main_test

import (
	"context"
	"net/http"
	"os"

	"fmt"

	"io/ioutil"
	"path"

	"code.cloudfoundry.org/cc-to-perm-migrator/capi/capimodels"
	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/retriever"
	permgofakes "code.cloudfoundry.org/perm/pkg/api/protos/protosfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"code.cloudfoundry.org/perm/pkg/api/protos"
	"github.com/onsi/gomega/ghttp"
)

const configTemplate = `log_level: info

uaa:
  url: %s
  ca_cert_path: %s

cloud_controller:
  url: %s
  client_id: perm-migrator
  client_secret: secret
  client_scopes: ["cloud_controller.admin_read_only"]
  ca_cert_path: %s

perm:
  hostname: %s
  port: %d
`

const ca = `-----BEGIN CERTIFICATE-----
MIIDMTCCAhmgAwIBAgIUR7aIygXu6VhofEraEgca2J4p5HcwDQYJKoZIhvcNAQEL
BQAwETEPMA0GA1UEAxMGVGVzdENBMB4XDTE4MDMyODE5MjEwOVoXDTE5MDMyODE5
MjEwOVowETEPMA0GA1UEAxMGVGVzdENBMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAu/1tGXsIurUcm1lUhXqmtXcw+GnRpyPiK7I9OQo+8A486xAy5n/W
s82CBau7IT3ZoOeBlI+OdHGjuA1ZiQ0KL8xBTbtqJ2nTh2HFjhp+4BqPjCeYWgev
J2DbIV1PdTAWs4HsdHGbEQWOupxnR+2mtYeGWSnyfGweMpXMW+EYKLinRDGt8wVB
kRpJ/LzL26VGeDkAi3Qofqj9EtrZ7z0/F+OhuMpDdBti2jehoz6t6BRvioW+tQ9i
lk5XBUSYE/pkoI2ZBbVKkRvlhO5GyIE0nOZ3KlsGwJghpx94aKfowsEq6H5+Jp8T
lelQcTobfxOMZ4SQhXBqh2P9uvpHZCe0qQIDAQABo4GAMH4wHQYDVR0OBBYEFMXF
QyNEQw6DDCwacyECtdCKs5umMEwGA1UdIwRFMEOAFMXFQyNEQw6DDCwacyECtdCK
s5umoRWkEzARMQ8wDQYDVQQDEwZUZXN0Q0GCFEe2iMoF7ulYaHxK2hIHGtieKeR3
MA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBABSVepMJKKj1DNL1
/hy/S+oyC+rKr9Yi+0MXTJDVMA2WwVDn+76BH2KHOGleT+UnkesUy8Yj8+zSBDs5
d6GPStye64HPkrmURfB7IvOqRV9Pg1efeP28vXuXtsp9OKSsf6CGuP4daumExt+t
wzYLM7/KyXUHNbEb4dvd5zi6JQGxBvpAInRKrMioj+rz9z8sizBokZPWS4jlm1YT
0rXiZwoZHDBK/PWtVp8WittYjZ2Whe873rhkJl9gFkOf5S0UWYJgSFFoCC647lD7
hGnHthGDF1mb/w+/sQJ/PwjOUFKgH8chzCby4US28yCiZoe4AbgI+6ksBQVj4Zve
MEl+WMg=
-----END CERTIFICATE-----`

var _ = Describe("CCToPermMigrator", func() {
	var (
		roleServiceServer *permgofakes.FakeRoleServiceServer

		ccServer   *ghttp.Server
		permServer *PermServer

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

		numAssignments int
	)

	BeforeEach(func() {
		roleServiceServer = new(permgofakes.FakeRoleServiceServer)

		var err error

		ccServer = ghttp.NewServer()
		permServer, err = NewPermServer(roleServiceServer)

		Expect(err).NotTo(HaveOccurred())

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

		roleServiceServer.CreateRoleStub = func(ctx context.Context, req *protos.CreateRoleRequest) (*protos.CreateRoleResponse, error) {
			return &protos.CreateRoleResponse{
				Role: &protos.Role{
					Name: req.GetName(),
				},
			}, nil
		}

		roleServiceServer.AssignRoleStub = func(ctx context.Context, req *protos.AssignRoleRequest) (*protos.AssignRoleResponse, error) {
			return &protos.AssignRoleResponse{}, nil
		}
	})

	AfterEach(func() {
		ccServer.Close()
		permServer.Stop()
	})

	Describe("#Main", func() {
		var configFilePath string
		var tmpDir string
		var err error
		var uaaCAPath, ccCAPath string
		BeforeEach(func() {
			//These handlers are appended in the order in which they are called.
			//If adding more handlers, make sure they are placed correctly in the set of calls.
			appendHandler(ccServer, "GET", fmt.Sprintf("/oauth/token%s", retriever.OpenIDConfigurationEndpoint), new(interface{}))
			appendHandler(ccServer, "POST", "/oauth/token", tokenJSON{
				AccessToken:  "cool",
				TokenType:    "whatever",
				RefreshToken: "something",
				ExpiresIn:    "1234",
			})
			appendHandler(ccServer, "GET", "/v2/organizations", orgsPage1)
			appendHandler(ccServer, "GET", orgsNextURL, orgsPage2)
			appendHandler(ccServer, "GET", "/v2/organizations/org-guid-1/user_roles", orgRoles1Page1)
			appendHandler(ccServer, "GET", orgRolesNextURL, orgRoles1Page2)
			appendHandler(ccServer, "GET", "/v2/organizations/org-guid-1/spaces", spaces1Page1)
			appendHandler(ccServer, "GET", spacesNextURL, spaces1Page2)
			appendHandler(ccServer, "GET", "/v2/spaces/space-guid-1/user_roles", spaceRoles1Page1)
			appendHandler(ccServer, "GET", spaceRolesNextURL, spaceRoles1Page2)

			appendHandler(ccServer, "GET", "/v2/spaces/space-guid-2/user_roles", spaceRoles2)
			appendHandler(ccServer, "GET", "/v2/organizations/org-guid-2/user_roles", orgRoles2)
			appendHandler(ccServer, "GET", "/v2/organizations/org-guid-2/spaces", spaces2)
			appendHandler(ccServer, "GET", "/v2/spaces/space-guid-3/user_roles", spaceRoles3)

			tmpDir, err = ioutil.TempDir("", "ccmtest")
			Expect(err).NotTo(HaveOccurred())
			configFilePath = path.Join(tmpDir, "config.yml")

			uaaCAPath = path.Join(tmpDir, "uaa-ca.cert")
			err = ioutil.WriteFile(uaaCAPath, []byte(ca), 0600)
			Expect(err).NotTo(HaveOccurred())

			ccCAPath = path.Join(tmpDir, "cc-ca.cert")
			err = ioutil.WriteFile(ccCAPath, []byte(ca), 0600)
			Expect(err).NotTo(HaveOccurred())

			contents := fmt.Sprintf(configTemplate, ccServer.URL(), uaaCAPath, ccServer.URL(), ccCAPath, permServer.Hostname(), permServer.Port())

			err = ioutil.WriteFile(configFilePath, []byte(contents), 0600)
			Expect(err).NotTo(HaveOccurred())

			numAssignments = 0
			for _, resource := range orgRoles1Page1.Resources {
				for i := 0; i < len(resource.Entity.Roles); i++ {
					numAssignments++
				}
			}
			for _, resource := range orgRoles1Page2.Resources {
				for i := 0; i < len(resource.Entity.Roles); i++ {
					numAssignments++
				}
			}
			for _, resource := range orgRoles2.Resources {
				for i := 0; i < len(resource.Entity.Roles); i++ {
					numAssignments++
				}
			}
			for _, resource := range spaceRoles1Page1.Resources {
				for i := 0; i < len(resource.Entity.Roles); i++ {
					numAssignments++
				}
			}
			for _, resource := range spaceRoles1Page2.Resources {
				for i := 0; i < len(resource.Entity.Roles); i++ {
					numAssignments++
				}
			}
			for _, resource := range spaceRoles2.Resources {
				for i := 0; i < len(resource.Entity.Roles); i++ {
					numAssignments++
				}
			}
			for _, resource := range spaceRoles3.Resources {
				for i := 0; i < len(resource.Entity.Roles); i++ {
					numAssignments++
				}
			}
		})

		AfterEach(func() {
			err := os.RemoveAll(tmpDir)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the UAA CA certificate is unusable", func() {
			BeforeEach(func() {
				ccServer.AllowUnhandledRequests = true
			})

			It("fails fast when the cert is invalid", func() {
				uaaCAPath := path.Join(tmpDir, "uaa-ca.cert")
				err = ioutil.WriteFile(uaaCAPath, []byte(`invalid`), 0600)
				Expect(err).NotTo(HaveOccurred())
				contents := fmt.Sprintf(configTemplate, ccServer.URL(), uaaCAPath, ccServer.URL(), ccCAPath, permServer.Hostname(), permServer.Port())
				err = ioutil.WriteFile(configFilePath, []byte(contents), 0600)

				Expect(err).NotTo(HaveOccurred())
				session := RunCommand("--config-file-path", configFilePath)
				Eventually(session).Should(gexec.Exit(1))
			})

			It("fails fast when the cert cannot be read", func() {
				contents := fmt.Sprintf(configTemplate, ccServer.URL(), tmpDir, ccServer.URL(), ccCAPath, permServer.Hostname(), permServer.Port())
				err = ioutil.WriteFile(configFilePath, []byte(contents), 0600)
				Expect(err).NotTo(HaveOccurred())
				session := RunCommand("--config-file-path", configFilePath)
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("when the CloudController's CA certificate is unusable", func() {
			BeforeEach(func() {
				ccServer.AllowUnhandledRequests = true
			})

			It("fails fast when the cert is invalid", func() {
				ccCAPath := path.Join(tmpDir, "cc-ca.cert")
				err = ioutil.WriteFile(ccCAPath, []byte(`invalid`), 0600)
				Expect(err).NotTo(HaveOccurred())
				contents := fmt.Sprintf(configTemplate, ccServer.URL(), uaaCAPath, ccServer.URL(), ccCAPath, permServer.Hostname(), permServer.Port())
				err = ioutil.WriteFile(configFilePath, []byte(contents), 0600)

				Expect(err).NotTo(HaveOccurred())
				session := RunCommand("--config-file-path", configFilePath)
				Eventually(session).Should(gexec.Exit(1))
			})

			It("fails fast when the cert cannot be read", func() {
				contents := fmt.Sprintf(configTemplate, ccServer.URL(), uaaCAPath, ccServer.URL(), tmpDir, permServer.Hostname(), permServer.Port())
				err = ioutil.WriteFile(configFilePath, []byte(contents), 0600)
				Expect(err).NotTo(HaveOccurred())
				session := RunCommand("--config-file-path", configFilePath)
				Eventually(session).Should(gexec.Exit(1))
			})

			It("doesn't fail when the no CA certs for CC are supplied", func() {
				err = ioutil.WriteFile(ccCAPath, []byte("\n"), 0600)
				Expect(err).NotTo(HaveOccurred())
				contents := fmt.Sprintf(configTemplate, ccServer.URL(), uaaCAPath, ccServer.URL(), ccCAPath, permServer.Hostname(), permServer.Port())
				err = ioutil.WriteFile(configFilePath, []byte(contents), 0600)
				Expect(err).NotTo(HaveOccurred())
				session := RunCommand("--config-file-path", configFilePath)
				Eventually(session).Should(gexec.Exit(0))
			})
		})

		Context("when the config flag is not passed", func() {
			BeforeEach(func() {
				ccServer.AllowUnhandledRequests = true
			})

			It("exits with 1", func() {
				session := RunCommand()
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("when getting the issuer from the OIDC provider fails", func() {
			BeforeEach(func() {
				ccServer.SetHandler(0,
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/oauth/token%s", retriever.OpenIDConfigurationEndpoint)),
						ghttp.RespondWithJSONEncoded(http.StatusNotFound, `{}`),
					),
				)
			})

			It("exits with 1", func() {
				session := RunCommand("--config-file-path", configFilePath)
				Eventually(session.Out).Should(gbytes.Say("failed-to-get-issuer-from-oidc-provider"))
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("in regular (non-dry-mode) mode", func() {
			BeforeEach(func() {
				f, err := os.OpenFile(configFilePath, os.O_APPEND|os.O_WRONLY, 0644)
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()

				_, err = f.WriteString("\ndry_run: false\n")
				Expect(err).NotTo(HaveOccurred())
			})

			It("runs successfully", func() {
				session := RunCommand("--config-file-path", configFilePath)

				Eventually(session).Should(gexec.Exit(0))

				Eventually(session.Err).Should(gbytes.Say("Number of role assignments: %d", numAssignments))
				Eventually(session.Err).Should(gbytes.Say("Total errors: 0"))
			})
		})

		Context("in dry-run mode", func() {
			BeforeEach(func() {
				f, err := os.OpenFile(configFilePath, os.O_APPEND|os.O_WRONLY, 0644)
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()

				_, err = f.WriteString("\ndry_run: true\n")
				Expect(err).NotTo(HaveOccurred())
			})

			It("runs successfully", func() {
				session := RunCommand("--config-file-path", configFilePath)

				Eventually(session).Should(gexec.Exit(0))

				Eventually(session.Err).Should(gbytes.Say("DRY-RUN; ROLE ASSIGNMENTS WILL NOT BE MIGRATED"))
				Eventually(session.Err).Should(gbytes.Say("Number of role assignments: %d", numAssignments))
				Eventually(session.Err).Should(gbytes.Say("Total errors: 0"))
				Eventually(session.Err).Should(gbytes.Say("DRY-RUN; ROLE ASSIGNMENTS WERE NOT MIGRATED"))
			})
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
