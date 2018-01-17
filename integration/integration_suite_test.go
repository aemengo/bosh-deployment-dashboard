package integration

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
	"testing"
	"github.com/aemengo/bosh-deployment-dashboard/config"
	"gopkg.in/yaml.v2"
	"os/exec"
	"io/ioutil"
)

var (
	agentBinaryPath string
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = BeforeSuite(func() {
	var err error
	agentBinaryPath, err = gexec.Build("github.com/aemengo/bosh-deployment-dashboard/cmd/agent")
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
