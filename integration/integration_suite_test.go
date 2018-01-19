package integration

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aemengo/bosh-deployment-dashboard/config"
	"github.com/onsi/gomega/gexec"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
	"testing"
	"github.com/aemengo/bosh-deployment-dashboard/info"
	"encoding/json"
	"fmt"
	"net/http"
	"bytes"
	"github.com/jmoiron/sqlx"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"net"
	"github.com/onsi/gomega/gbytes"
)

var (
	agentBinaryPath string
	hubBinaryPath   string
	hubPort         = "4567"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Second)

	var err error
	agentBinaryPath, err = gexec.Build("github.com/aemengo/bosh-deployment-dashboard/cmd/agent")
	Expect(err).NotTo(HaveOccurred())

	hubBinaryPath, err = gexec.Build("github.com/aemengo/bosh-deployment-dashboard/cmd/hub")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func StartAgentWithConfig(cfg config.Config) *gexec.Session {
	contents, _ := yaml.Marshal(cfg)
	ioutil.WriteFile("/tmp/bdd-agent-test-config.yml", contents, 0600)
	cmd := exec.Command(agentBinaryPath, "/tmp/bdd-agent-test-config.yml")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

func StartHubWithConfig(cfg config.Config) *gexec.Session {
	contents, _ := yaml.Marshal(cfg)
	ioutil.WriteFile("/tmp/bdd-hub-test-config.yml", contents, 0600)
	cmd := exec.Command(hubBinaryPath, "/tmp/bdd-hub-test-config.yml")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	//wait for server to create database tables
	Eventually(session).Should(gbytes.Say(`Initializing hub on addr`))

	Eventually(func() error {
		_, err := net.DialTimeout("tcp","127.0.0.1:"+hubPort, time.Second)
		return err
	}, 10*time.Second, time.Second).ShouldNot(HaveOccurred(), "test hub not reachable in time")

	return session
}

func PostHub(path string, body info.Info) *http.Response {
	contents, _ := json.Marshal(body)
	url := fmt.Sprintf("http://127.0.0.1:%s%s", hubPort, path)
	response, err := http.Post(url, "application/json", bytes.NewReader(contents))
	Expect(err).NotTo(HaveOccurred())
	return response
}

func HubGet(path string) *http.Response {
	url := fmt.Sprintf("http://127.0.0.1:%s%s", hubPort, path)
	response, err := http.Get(url)
	Expect(err).NotTo(HaveOccurred())
	return response
}

func GetDBClient(dataDir string) *sqlx.DB {
	db, err := sql.Open("sqlite3", dataDir+"/bdd-hub.db")
	Expect(err).NotTo(HaveOccurred())
	return sqlx.NewDb(db, "sqlite3")
}
