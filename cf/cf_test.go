package cf_test

import (
	"github.com/aemengo/bosh-deployment-dashboard/cf"

	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"net/http"
)

var _ = Describe("Cf", func() {

	var (
		server *ghttp.Server
		cfInfo cf.CF
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		cfInfo = cf.CF{
			ApiURL: server.URL(),
		}
	})

	AfterEach(func() {
		server.Close()
	})

	It("works", func() {
		bindingsContents, _ := json.Marshal(cf.Response{
			Resources: []cf.Resource{{
				Entity: cf.Entity{
					"app_guid": "some-app-guid",
				},
			}},
		})

		appContents, _ := json.Marshal(cf.EntityResponse{
			Entity: cf.Entity{
				"name": "beautiful-app-name",
				"space_guid": "some-space-guid",
			},
		})

		spaceContents, _ := json.Marshal(cf.EntityResponse{
			Entity: cf.Entity{
				"name": "beautiful-space-name",
				"organization_guid": "some-organization-guid",
			},
		})

		orgContents, _ := json.Marshal(cf.EntityResponse{
			Entity: cf.Entity{
				"name": "beautiful-organization-name",
			},
		})

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/v2/service_instances/some-guid/service_bindings"),
				ghttp.RespondWith(http.StatusOK, bindingsContents),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/v2/apps/some-app-guid"),
				ghttp.RespondWith(http.StatusOK, appContents),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid"),
				ghttp.RespondWith(http.StatusOK, spaceContents),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/v2/organizations/some-organization-guid"),
				ghttp.RespondWith(http.StatusOK, orgContents),
			),
		)

		info, _ := cfInfo.GetInfo("some-guid")
		Expect(info.OrgName).To(Equal("beautiful-organization-name"))
		Expect(info.SpaceName).To(Equal("beautiful-space-name"))
		Expect(info.BoundAppNames).To(Equal([]string{"beautiful-app-name"}))
	})
})
