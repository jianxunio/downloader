package config

import "os"
import "fmt"
import "encoding/json"

type SSDBNode struct {
	Host string
	Port int
}

type Config struct {
	RabbitmqIp          string
	RabbitmqPort        int
	RabbitmqManagerIp   string
	RabbitmqManagerPort int
	RabbitmqUser        string
	RabbitmqPassword    string
	RabbitmqConsumers   int
	ProxyRedisIp        string
	ProxyRedisPort      int
	SSDBNodes           []SSDBNode
}

type ProxyConfigNode struct {
	AddTimeout     int
	AddFrequency   int
	Name           string
	ProxyUrl       string
	Status         int
	CheckTimeout   int
	CheckFrequency int
	CheckUrl       string
	CheckRequests  int
	ProxyMax       int // When proxy number come to this limit, AddProxyWorker will not add proxy, 0 means unlimited.
}

type ProxyConfig struct {
	ProxyRedisIp   string
	ProxyRedisPort int
	Configs        []ProxyConfigNode
}

const (
	StatusStopped = 0
	StatusRunning = 1
)

func loadFile(filePath string) (*os.File, error) {
	var file *os.File
	var err error
	_, err = os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	file, err = os.Open(filePath)
	return file, err
}

func GetProxyConfig(params ...string) (ProxyConfig, error) {
	var filePath string
	if len(params) > 0 {
		filePath = params[0]
	} else {
		filePath = "/etc/yascrapy/proxy.json"
	}
	cfg := ProxyConfig{}
	file, err := loadFile(filePath)
	if err != nil {
		fmt.Printf("[error] read file `proxy.json` fail:%s\n", err.Error())
		return cfg, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	return cfg, err
}

func Get(params ...string) (Config, error) {
	var filePath string
	if len(params) > 0 {
		filePath = params[0]
	} else {
		filePath = "/etc/yascrapy/core.json"
	}
	cfg := Config{}
	file, err := loadFile(filePath)
	if err != nil {
		return cfg, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
