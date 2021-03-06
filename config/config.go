package config

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Spec struct {
	ID           string `yaml:"id" json:"id"`
	InstanceName string `yaml:"instance_name" json:"instance_name"`
	Address      string `yaml:"address" json:"address"`
	AZ           string `yaml:"az" json:"az"`
	Deployment   string `yaml:"deployment" json:"deployment"`
	Index        int    `yaml:"index" json:"index"`
	IP           string `yaml:"ip" json:"ip"`
}

type Hub struct {
	IP      string `yaml:"ip"`
	Port    string `yaml:"port"`
	DataDir string `yaml:"data_dir"`
	WebDir  string `yaml:"web_dir"`
}

type Config struct {
	Spec  Spec   `yaml:"spec"`
	Hub   Hub    `yaml:"hub"`
	Label string `yaml:"label"`
}

func NewConfig(path string) (Config, error) {
	configContents, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, errors.Wrap(err, "unable to find config file at path: "+path)
	}

	var cfg Config

	err = yaml.Unmarshal(configContents, &cfg)
	if err != nil {
		return Config{}, errors.Wrap(err, "unable to read config file")
	}

	return cfg, nil
}

func (h *Hub) Addr() string {
	return h.IP+":"+h.Port
}