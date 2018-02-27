package cf

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

type DeploymentInfo struct {
	AppNames  []string
	SpaceName string
	OrgName   string
}

type CFEntity struct {
	Name string `json:"name"`
}

type CFResource struct {
	Entity CFEntity `json:"entity"`
}

type CFSpaceEntity struct {
	CFEntity
	OrganizationGuid string `json:"organization_guid"`
}

type CFSpaceResource struct {
	Entity CFSpaceEntity `json:"entity"`
}

type CFAppEntity struct {
	Name      string `json:"name"`
	SpaceGuid string `json:"space_guid"`
}

type CFAppResource struct {
	Entity CFAppEntity `json:"entity"`
}

type CFServiceBindingEntity struct {
	AppGuid string `json:"app_guid"`
}

type CFServiceBindingResource struct {
	Entity CFServiceBindingEntity `json:"entity"`
}

type CFServiceBindings struct {
	Resources []CFServiceBindingResource `json:"resources"`
}

type CF struct {
	URL string
}

var (
	deploymentNameRe = regexp.MustCompile("service-instance_(.*)")
)

func (cf *CF) GetDeploymentInfo(deploymentName string) (DeploymentInfo, error) {
	matches := deploymentNameRe.FindStringSubmatch(deploymentName)
	if matches == nil {
		return DeploymentInfo{}, nil
	}

	var (
		serviceInstanceId  = matches[1]
		serviceBindingsURL = fmt.Sprintf("/v2/service_instances/%s/service_bindings", serviceInstanceId)
		serviceBindings    = CFServiceBindings{}
		deploymentInfo     = DeploymentInfo{}
		spaceGUID          = ""
	)

	err := cf.makeRequest(serviceBindingsURL, &serviceBindings)
	if err != nil {
		return DeploymentInfo{}, err
	}

	for _, resource := range serviceBindings.Resources {
		var (
			appGuid     = resource.Entity.AppGuid
			appURL      = fmt.Sprintf("/v2/apps/%s", appGuid)
			appResource = CFAppResource{}
		)

		cf.makeRequest(appURL, &appResource)
		deploymentInfo.AppNames = append(deploymentInfo.AppNames, appResource.Entity.Name)
		spaceGUID = appResource.Entity.SpaceGuid
	}

	if spaceGUID == "" {
		return deploymentInfo, nil
	}

	var spaceResource CFSpaceResource
	cf.makeRequest("/v2/spaces/"+spaceGUID, &spaceResource)
	deploymentInfo.SpaceName = spaceResource.Entity.Name

	var organizationResource CFResource
	cf.makeRequest("/v2/organizations/"+spaceResource.Entity.OrganizationGuid, &organizationResource)
	deploymentInfo.OrgName = organizationResource.Entity.Name

	return deploymentInfo, nil
}

func (cf *CF) makeRequest(path string, dest interface{}) error {
	resp, _ := http.Get(cf.URL + path)

	switch resp.StatusCode {
	case http.StatusOK:
		json.NewDecoder(resp.Body).Decode(dest)
	case http.StatusNotFound:
		return fmt.Errorf("invalid response [%s] for GET %s", resp.Status, path)
	}

	return nil
}
