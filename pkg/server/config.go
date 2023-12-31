package server

import (
	"io/ioutil"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gopkg.in/yaml.v2"
)

type Targets struct {
	ProberType string   `yaml:"prober_type"`
	Region     string   `yaml:"region"`
	Target     []string `yaml:"target"`
}
type Config struct {
	RpcListenAddr     string     `yaml:"rpc_listen_addr"`
	MetricsListenAddr string     `yaml:"metrics_listen_addr"`
	ProberTargets     []*Targets `yaml:"prober_targets"`
	PointNum          int        `yaml:"point_num"`
	NodeName          string     `yaml:"node_name"`
}

var PointNum int

func Load(s string) (*Config, error) {
	cfg := &Config{}

	err := yaml.UnmarshalStrict([]byte(s), cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func LoadFile(filename string, logger log.Logger) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg, err := Load(string(content))
	if err != nil {
		level.Error(logger).Log("msg", "parsing YAML file errr...", "error", err)
	}
	return cfg, nil
}
