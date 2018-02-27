package cf_test

import (
	"github.com/aemengo/bosh-deployment-dashboard/cf"

	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("CF", func() {

	var (
		client         cf.CF
		deploymentName = "service-instance_deployment-id"
		server         *ghttp.Server
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		client = cf.CF{
			URL: server.URL(),
		}
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("GetDeploymentInfo", func() {
		Context("when the given deployment name doesn't have 'service-instance_'", func() {
			It("makes no API calls and returns no deployment info", func() {
				deploymentInfo, err := client.GetDeploymentInfo("some-other-type-of-deployment-id")

				Expect(err).NotTo(HaveOccurred())
				Expect(deploymentInfo).To(Equal(cf.DeploymentInfo{}))
			})
		})

		Context("when no service instance exists for the given deployment name", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/service_instances/deployment-id/service_bindings"),
						ghttp.RespondWith(http.StatusNotFound, nil)),
				)
			})

			It("returns no appNames", func() {
				_, err := client.GetDeploymentInfo(deploymentName)

				Expect(err).To(MatchError("invalid response [404 Not Found] for GET /v2/service_instances/deployment-id/service_bindings"))
			})
		})

		Context("when there are no apps for the given deployment name", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/service_instances/deployment-id/service_bindings"),
						ghttp.RespondWith(http.StatusOK, `{
					      "resources":[]
					    }`)),
				)
			})

			It("returns no appNames", func() {
				deploymentInfo, err := client.GetDeploymentInfo(deploymentName)

				Expect(err).NotTo(HaveOccurred())
				Expect(deploymentInfo.AppNames).To(BeEmpty())
			})
		})

		Context("when there are apps bound for the given deployment name", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/service_instances/deployment-id/service_bindings"),
						ghttp.RespondWith(http.StatusOK, `{
					      "resources":[{ "entity" : { "app_guid" : "app-guid-1" } }]
					    }`),
					),

					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/apps/app-guid-1"),
						ghttp.RespondWith(http.StatusOK, `{
						  "entity" : { "name": "app-name-1", "space_guid": "space-guid-1" }
						}`),
					),

					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodGet, "/v2/spaces/space-guid-1"),
						ghttp.RespondWith(http.StatusOK, `{
						  "entity" : { "name": "space-name", "organization_guid": "organization-guid-1" }
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

			It("returns the appropriate app names", func() {
				deploymentInfo, err := client.GetDeploymentInfo(deploymentName)

				Expect(err).NotTo(HaveOccurred())
				Expect(deploymentInfo.AppNames).To(Equal([]string{"app-name-1"}))
			})

			It("returns the appropriate space name", func() {
				deploymentInfo, err := client.GetDeploymentInfo(deploymentName)

				Expect(err).NotTo(HaveOccurred())
				Expect(deploymentInfo.SpaceName).To(Equal("space-name"))
			})

			It("returns the appropriate organization name", func() {
				deploymentInfo, err := client.GetDeploymentInfo(deploymentName)

				Expect(err).NotTo(HaveOccurred())
				Expect(deploymentInfo.OrgName).To(Equal("org-name"))
			})
		})
	})
})
