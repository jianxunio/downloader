package config

import "testing"

func TestGetConfig(t *testing.T) {
	cfg, err := Get("/etc/yascrapy/core.json")
	if err != nil {
		t.Error(err)
	}
	proxyNode := SSDBNode {
        Host: cfg.ProxyRedisIp,
        Port: cfg.ProxyRedisPort,
    }
	t.Log(cfg)
	t.Log(cfg.SSDBNodes)
	t.Log(proxyNode)
}

func TestGetProxyConfig(t *testing.T) {
	cfg, err := GetProxyConfig("/etc/yascrapy/proxy.json")
	if err != nil {
		t.Error(err)
	}
	t.Log(cfg)
	t.Log(cfg.ProxyRedisPort)
	t.Log(cfg.ProxyRedisIp)
	t.Log(cfg.Configs)
}
