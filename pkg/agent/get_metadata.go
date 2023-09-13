package agent

import (
	"fmt"
	"io/ioutil"
	"net"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gopkg.in/yaml.v2"
)

var (
	LocalRegion string

	LocalIp       string
	AgentRound    int64
	LocalNodeName string
	LocalVLANID   int64
	LocalZid      int
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

func GetLocalIp(logger log.Logger) bool {

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		level.Error(logger).Log("msg", "GetLocalIp_net.InterfaceAddrs", "err", err)
		return false
	}

	for _, addr := range addrs {

		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() == nil {
				addr := ipnet.IP.String()
				LocalIp = addr

				return true
			}
		}
	}
	return false

}

func GetLocalVLANID(logger log.Logger) bool {
	LocalVLANID = AgentRound
	return true
}

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

type PointInfo struct {
	Zid   int
	YwNic string
}

func GetLocalNodeName(logger log.Logger) bool {

	configFile := "../../../prober.yml"
	sConfig, err := LoadFile(configFile, logger)
	if err != nil {
		level.Error(logger).Log("msg", "load_config_file_error", "err", err)
		return false
	}
	LocalNodeName = sConfig.NodeName

	LocalZid, err = strconv.Atoi(LocalNodeName)
	if err != nil {
		fmt.Println("get_metadata.go:String conversion to int failed.", err)
		return false
	}
	LocalZid = (LocalZid % 40) + 1
	LocalRegion = strconv.Itoa(LocalZid)
	return true
}
