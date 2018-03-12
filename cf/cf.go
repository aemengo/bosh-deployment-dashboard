package cf

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/aemengo/bosh-deployment-dashboard/config"

	"context"

	"encoding/json"

	"io/ioutil"

	"github.com/pkg/errors"
	clientcredential "golang.org/x/oauth2/clientcredentials"
)

type DeploymentInfo struct {
	AppNames  []string
	SpaceName string
	OrgName   string
}

type Cf struct {
	cfg        config.Config
	httpClient *http.Client
}

func New(cfg config.Config) *Cf {
	ccCfg := &clientcredential.Config{
		ClientID:     cfg.Cf.ClientID,
		ClientSecret: cfg.Cf.ClientSecret,
		TokenURL:     cfg.Cf.UaaHost,
	}

	httpClient := ccCfg.Client(context.Background())

	return &Cf{
		cfg:        cfg,
		httpClient: httpClient,
	}
}

var (
	deploymentNameRe = regexp.MustCompile("service-instance_(.*)")
)

func (c *Cf) GetDeploymentInfo() (DeploymentInfo, error) {
	deploymentName := c.cfg.Spec.Deployment

	matches := deploymentNameRe.FindStringSubmatch(deploymentName)
	if matches == nil {
		return DeploymentInfo{}, errors.Errorf("the following deployment name does not match the pattern of the on-demand-service-broker: %s", deploymentName)
	}

	var (
		serviceInstanceId   = matches[1]
		serviceInstanceResp = struct {
			Entity struct {
				SpaceURL           string `json:"space_url"`
				ServiceBindingsURL string `json:"service_bindings_url"`
			} `json:"entity"`
		}{}
	)

	err := c.makeRequest("/v2/service_instances/"+serviceInstanceId, &serviceInstanceResp)
	if err != nil {
		return DeploymentInfo{}, err
	}

	spaceResp := struct {
		Entity struct {
			Name            string `json:"name"`
			OrganizationURL string `json:"organization_url"`
		} `json:"entity"`
	}{}

	if err := c.makeRequest(serviceInstanceResp.Entity.SpaceURL, &spaceResp); err != nil {
		return DeploymentInfo{}, err
	}

	orgResp := struct {
		Entity struct {
			Name string `json:"name"`
		} `json:"entity"`
	}{}

	if err := c.makeRequest(spaceResp.Entity.OrganizationURL, &orgResp); err != nil {
		return DeploymentInfo{}, err
	}

	serviceBindingsResp := struct {
		Resources []struct {
			Entity struct {
				AppURL string `json:"app_url"`
			}
		}
	}{}

	if err := c.makeRequest(serviceInstanceResp.Entity.ServiceBindingsURL, &serviceBindingsResp); err != nil {
		return DeploymentInfo{}, err
	}

	var appNames []string
	for _, resource := range serviceBindingsResp.Resources {
		appResp := struct {
			Entity struct {
				Name string `json:"name"`
			} `json:"entity"`
		}{}

		if err := c.makeRequest(resource.Entity.AppURL, &appResp); err != nil {
			return DeploymentInfo{}, err
		}

		appNames = append(appNames, appResp.Entity.Name)
	}

	return DeploymentInfo{
		SpaceName: spaceResp.Entity.Name,
		OrgName:   orgResp.Entity.Name,
		AppNames:  appNames,
	}, nil
}

func (c *Cf) makeRequest(path string, dest interface{}) error {
	resp, err := http.Get(c.cfg.Cf.ApiHost + path)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		json.NewDecoder(resp.Body).Decode(dest)
	default:
		var contents []byte
		if resp.Body != nil {
			contents, _ = ioutil.ReadAll(resp.Body)
		}
		return fmt.Errorf("invalid response [%s] for GET %s: %s", resp.Status, path, contents)
	}

	return nil
}
