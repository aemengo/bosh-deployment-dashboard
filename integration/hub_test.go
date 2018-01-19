package integration

import (
	"github.com/aemengo/bosh-deployment-dashboard/config"
	"github.com/aemengo/bosh-deployment-dashboard/info"
	"github.com/aemengo/bosh-deployment-dashboard/system"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"net/http"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"os"
)

var _ = Describe("BDD Hub", func() {

	var (
		hubSession *gexec.Session
		cfg        config.Config
		dataDir    string
		systemInfo info.Info
		sqlxClient *sqlx.DB
	)

	BeforeEach(func() {
		var err error
		dataDir, err = ioutil.TempDir("", "bdd-hub-")
		Expect(err).NotTo(HaveOccurred())

		cfg = config.Config{
			HubAddr:    "127.0.0.1:" + hubPort,
			HubDataDir: dataDir,
		}

		systemInfo = info.Info{
			Spec: config.Spec{
				ID:         "some-id",
				Deployment: "some-deployment",
			},
			Label: "some-label",
			Stats: system.Stats{
				PersistentDiskUsed: 60,
			},
		}
	})

	AfterEach(func() {
		sqlxClient.Close()
		hubSession.Kill()
		os.RemoveAll(dataDir)
	})

	It("POST /health inserts and updates health metrics to hub", func() {
		hubSession = StartHubWithConfig(cfg)
		response := PostHub("/health", systemInfo)
		Expect(response.StatusCode).To(Equal(http.StatusOK))

		sqlxClient = GetDBClient(dataDir)

		var count int
		err := sqlxClient.Get(&count, "select count(*) from metrics")
		Expect(err).NotTo(HaveOccurred())
		Expect(count).To(Equal(1))

		var persistentDiskUsed int
		err = sqlxClient.Get(&persistentDiskUsed, "select persistent_disk_used from metrics limit 1")
		Expect(err).NotTo(HaveOccurred())
		Expect(persistentDiskUsed).To(Equal(60))

		systemInfo.Stats.PersistentDiskUsed = 80
		response = PostHub("/health", systemInfo)
		Expect(response.StatusCode).To(Equal(http.StatusOK))

		err = sqlxClient.Get(&count, "select count(*) from metrics")
		Expect(err).NotTo(HaveOccurred())
		Expect(count).To(Equal(1))

		err = sqlxClient.Get(&persistentDiskUsed, "select persistent_disk_used from metrics limit 1")
		Expect(err).NotTo(HaveOccurred())
		Expect(persistentDiskUsed).To(Equal(80))
	})

	It("GET /health returns the metrics saved", func() {
		systemInfo.Stats = system.Stats{
			PersistentDiskUsed: 30,
		}

		hubSession = StartHubWithConfig(cfg)
		response := PostHub("/health", systemInfo)
		Expect(response.StatusCode).To(Equal(http.StatusOK))

		response = HubGet("/health")
		Expect(response.StatusCode).To(Equal(http.StatusOK))

		contents, err := ioutil.ReadAll(response.Body)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(contents)).Should(SatisfyAll(
			ContainSubstring(`"deployment":"some-deployment`),
			ContainSubstring(`"label":"some-label"`),
			ContainSubstring(`"persistent_disk_used":30`),
		))
	})
})
