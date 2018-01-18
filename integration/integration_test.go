package integration

import (
	"github.com/aemengo/bosh-deployment-dashboard/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
)

var _ = Describe("BDD Agent", func() {

	var (
		agentSession      *gexec.Session
		server            *httptest.Server
		cfg               config.Config
		actualRequestBody string
	)

	BeforeEach(func() {
		actualRequestBody = ""

		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(_ http.ResponseWriter, r *http.Request) {
			contents, _ := ioutil.ReadAll(r.Body)
			actualRequestBody = string(contents)
		})
		server = httptest.NewServer(mux)
		u, _ := url.Parse(server.URL)

		cfg = config.Config{
			Spec: config.Spec{
				Deployment: "some-deployment-name",
			},
			HubAddr: u.Host,
			Label:   "some-deployment-type",
		}
	})

	AfterEach(func() {
		agentSession.Kill()
		server.Close()
	})

	It("sends health metrics to hub", func() {
		agentSession = StartAgentWithConfig(cfg)

		Eventually(func() string {
			return actualRequestBody
		}, "20s").Should(SatisfyAll(
			ContainSubstring(`"deployment":"some-deployment-name`),
			ContainSubstring(`"label":"some-deployment-type"`),
			ContainSubstring(`"system_stats":`),
		))
	})
})
