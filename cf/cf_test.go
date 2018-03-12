package cf_test

import (
	"github.com/aemengo/bosh-deployment-dashboard/cf"

	"net/http"

	"github.com/aemengo/bosh-deployment-dashboard/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("CF", func() {

	var (
		client *cf.Cf
		cfg    config.Config
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		cfg = config.Config{
			Cf: config.Cf{
				ApiHost: server.URL(),
			},
			Spec: config.Spec{
				Deployment: "service-instance_deployment-id",
			},
		}
	})

	AfterEach(func() {
		server.Close()
	})

	JustBeforeEach(func() {
		client = cf.New(cfg)
	})

	Describe("GetDeploymentInfo", func() {
		Context("when the given deployment name doesn't have 'service-instance_'", func() {
			BeforeEach(func() {
				cfg.Spec.Deployment = "some-bad-deployment-name"
			})

			It("makes no API calls and returns an error", func() {
				_, err := client.GetDeploymentInfo()
				Expect(err).To(MatchError("the following deployment name does not match the pattern of the on-demand-service-broker: some-bad-deployment-name"))
			})
		})

		Context("when an unexpected response code is returned", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/service_instances/deployment-id"),
						ghttp.RespondWith(http.StatusNotFound, `some-response`),
					),
				)
			})

			It("returns no appNames", func() {
				_, err := client.GetDeploymentInfo()
				Expect(err).To(MatchError("invalid response [404 Not Found] for GET /v2/service_instances/deployment-id: some-response"))
			})
		})

		Context("when valid responses are given", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/service_instances/deployment-id"),
						ghttp.RespondWith(http.StatusOK, `{
					      "entity" : { "space_url" : "/v2/spaces/some-space-guid", "service_bindings_url" : "/v2/service_instances/deployment-id/service_bindings" }
					    }`),
					),

					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid"),
						ghttp.RespondWith(http.StatusOK, `{
						  "entity" : { "name": "space-name", "organization_url": "/v2/organizations/organization-guid-1" }
						}`),
					),

					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/organizations/organization-guid-1"),
						ghttp.RespondWith(http.StatusOK, `{
						  "entity" : { "name": "org-name" }
						}`),
					),
				)
			})

			Context("when there are no applications bound", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest(http.MethodGet, "/v2/service_instances/deployment-id/service_bindings"),
							ghttp.RespondWith(http.StatusOK, `{
							"resources" : []
						}`),
						),
					)
				})

				It("returns the appropriate service deployment info", func() {
					deploymentInfo, err := client.GetDeploymentInfo()

					Expect(err).NotTo(HaveOccurred())
					Expect(deploymentInfo.SpaceName).To(Equal("space-name"))
					Expect(deploymentInfo.OrgName).To(Equal("org-name"))
					Expect(deploymentInfo.AppNames).To(BeEmpty())
				})
			})

			Context("when there are application bound to the instance", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest(http.MethodGet, "/v2/service_instances/deployment-id/service_bindings"),
							ghttp.RespondWith(http.StatusOK, `{	
								"resources" : [{"entity": { "app_url" : "/v2/apps/app-guid-1" }}]
							}`),
						),

						ghttp.CombineHandlers(
							ghttp.VerifyRequest(http.MethodGet, "/v2/apps/app-guid-1"),
							ghttp.RespondWith(http.StatusOK, `{
						  		"entity" : { "name": "app-name-1" }
							}`),
						),
					)
				})

				It("returns application info", func() {
					deploymentInfo, err := client.GetDeploymentInfo()

					Expect(err).NotTo(HaveOccurred())
					Expect(deploymentInfo.AppNames).To(Equal([]string{"app-name-1"}))
				})
			})
		})
	})
})
