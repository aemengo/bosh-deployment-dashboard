package info

import (
	"github.com/aemengo/bosh-deployment-dashboard/system"
	"github.com/aemengo/bosh-deployment-dashboard/config"
)

type Info struct {
	Spec  config.Spec  `json:"spec"`
	Label string       `json:"label"`
	Stats system.Stats `json:"system_stats"`
}
