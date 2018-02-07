package cf

import (
	"encoding/json"
	"net/http"
)

type CF struct {
	ApiURL string
}

type Info struct {
	OrgName       string
	SpaceName     string
	BoundAppNames []string
}


type Response struct {
	Resources []Resource `json:"resources"`
}

type Resource struct {
	Entity Entity `json:"entity"`
}

type Entity map[string]interface{}


type EntityResponse struct {
	Entity Entity `json:"entity"`
}


func (i *CF) GetAppInfo(appGUID string) (string, string) {
	response, err := http.Get(i.ApiURL + "/v2/apps/" + appGUID)

	if err != nil {
		panic(err)
	}

	var parsedResponse EntityResponse
	if err := json.NewDecoder(response.Body).Decode(&parsedResponse); err != nil {
		panic(err)
	}

	return parsedResponse.Entity["name"].(string), parsedResponse.Entity["space_guid"].(string)
}


func (i *CF) GetSpaceInfo(spaceGUID string) (string, string) {
	response, err := http.Get(i.ApiURL + "/v2/spaces/" + spaceGUID)

	if err != nil {
		panic(err)
	}

	var parsedResponse EntityResponse
	if err := json.NewDecoder(response.Body).Decode(&parsedResponse); err != nil {
		panic(err)
	}

	return parsedResponse.Entity["name"].(string), parsedResponse.Entity["organization_guid"].(string)
}


func (i *CF) GetOrganizationInfo(organizationGUID string) string {
	response, err := http.Get(i.ApiURL + "/v2/organizations/" + organizationGUID)

	if err != nil {
		panic(err)
	}

	var parsedResponse EntityResponse
	if err := json.NewDecoder(response.Body).Decode(&parsedResponse); err != nil {
		panic(err)
	}

	return parsedResponse.Entity["name"].(string)
}


func (i *CF) GetInfo(deploymentGUID string) (Info, error) {
	response, err := http.Get(i.ApiURL + "/v2/service_instances/" + deploymentGUID + "/service_bindings")
	var appGUIDs []string
	var result Info
	var spaceGUID string
	var organizationGUID string

	if err != nil {
		panic(err)
	}

	var parsedResponse Response
	if err := json.NewDecoder(response.Body).Decode(&parsedResponse); err != nil {
		panic(err)
	}

	for _, resource := range parsedResponse.Resources {
		appGUIDs = append(appGUIDs, resource.Entity["app_guid"].(string))
	}

	for _, appGUID := range appGUIDs {
		var appName string
		appName, spaceGUID = i.GetAppInfo(appGUID)
		result.BoundAppNames = append(result.BoundAppNames, appName)
	}

	result.SpaceName, organizationGUID = i.GetSpaceInfo(spaceGUID)
	result.OrgName = i.GetOrganizationInfo(organizationGUID)

	return result, nil
}
