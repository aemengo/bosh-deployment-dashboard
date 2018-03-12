package info

import (
	"github.com/aemengo/bosh-deployment-dashboard/cf"
	"github.com/aemengo/bosh-deployment-dashboard/config"
	"github.com/aemengo/bosh-deployment-dashboard/system"
)

type Info struct {
	Spec  config.Spec       `json:"spec"`
	Label string            `json:"label"`
	Stats system.Stats      `json:"system_stats"`
	Cf    cf.DeploymentInfo `json:"cf,omitempty"`
}
