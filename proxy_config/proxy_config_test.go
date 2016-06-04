package proxy_config

import "testing"
import "downloader/config"
import "downloader/ssdb"

func TestAdd(t *testing.T) {
	cfg, err := config.GetProxyConfig()
	if err != nil {
		t.Error(err.Error())
	}
	node := config.SSDBNode{
		Host: cfg.ProxyRedisIp,
		Port: cfg.ProxyRedisPort,
	}
	client := ssdb.GetProxyClient(node, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	cfgs := cfg.Configs
	for _, cfg := range cfgs {
		err = Add(client, cfg)
		if err != nil {
			t.Error("TestAdd fail: ", err.Error())
		}
	}
}

func TestGetProxyNames(t *testing.T) {
	cfg, err := config.GetProxyConfig()
	if err != nil {
		t.Error(err.Error())
	}
	node := config.SSDBNode{
		Host: cfg.ProxyRedisIp,
		Port: cfg.ProxyRedisPort,
	}
	client := ssdb.GetProxyClient(node, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	names, err := GetProxyNames(client)
	if err != nil {
		t.Error("TestGetProxyNames fail: ", err.Error())
	} else {
		t.Log(names)
	}
}

func TestGet(t *testing.T) {
	pcfg, err := config.GetProxyConfig()
	if err != nil {
		t.Error(err.Error())
	}
	node := config.SSDBNode{
		Host: pcfg.ProxyRedisIp,
		Port: pcfg.ProxyRedisPort,
	}
	client := ssdb.GetProxyClient(node, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	cfg, err := Get(client, "http_china")
	if err != nil {
		t.Error("TestGet fail: ", err.Error())
	} else {
		t.Log(cfg)
		t.Log("TestGet success")
	}
}

func TestDelete(t *testing.T) {
	pcfg, err := config.GetProxyConfig()
	if err != nil {
		t.Error(err.Error())
	}
	node := config.SSDBNode{
		Host: pcfg.ProxyRedisIp,
		Port: pcfg.ProxyRedisPort,
	}
	client := ssdb.GetProxyClient(node, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	err = Delete(client, "http_china")
	if err != nil {
		t.Error("TestDelete fail: ", err.Error())
	} else {
		t.Log("TestDelete success")
	}
}
