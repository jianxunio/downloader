package proxy

import "testing"
import "downloader/config"
import "downloader/ssdb"

func BenchmarkGetCount(b *testing.B) {
	cfg, err := config.Get("/etc/yascrapy/core.json")
	if err != nil {
		b.Error(err.Error())
	}
	node := config.SSDBNode{
		Host: cfg.ProxyRedisIp,
		Port: cfg.ProxyRedisPort,
	}
	client := ssdb.GetProxyClient(node, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	for i := 0; i < b.N; i++ {
		cnt, err := GetCount(client, "http_china")
		if err != nil {
			b.Error("TestGetCount fail ", err.Error())
		} else {
			b.Log(cnt)
		}
	}
}

func BenchmarkGetOne(b *testing.B) {
	cfg, err := config.Get("/etc/yascrapy/core.json")
	if err != nil {
		b.Error(err.Error())
	}
	node := config.SSDBNode{
		Host: cfg.ProxyRedisIp,
		Port: cfg.ProxyRedisPort,
	}
	client := ssdb.GetProxyClient(node, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	for i := 0; i < b.N; i++ {
		content, err := GetOne(client, "http_china")
		if err != nil {
			b.Error("TestGetOne fail ", err.Error())
		} else {
			b.Log(content)
		}
	}
}

func TestPush(t *testing.T) {
	cfg, err := config.Get("/etc/yascrapy/core.json")
	if err != nil {
		t.Error(err.Error())
	}
	node := config.SSDBNode{
		Host: cfg.ProxyRedisIp,
		Port: cfg.ProxyRedisPort,
	}
	client := ssdb.GetProxyClient(node, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	err = Push(client, "127.0.0.1:4001", "http_china")
	if err != nil {
		t.Error("TestPush fail ", err.Error())
	} else {
		t.Log("TestPush ok")
	}
}

func TestGet(t *testing.T) {
	cfg, err := config.Get("/etc/yascrapy/core.json")
	if err != nil {
		t.Error(err.Error())
	}
	node := config.SSDBNode{
		Host: cfg.ProxyRedisIp,
		Port: cfg.ProxyRedisPort,
	}
	client := ssdb.GetProxyClient(node, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	_, err = Get(client, "http_china")
	if err != nil {
		t.Error("TestGet fail ", err.Error())
	} else {
		t.Log("TestGet ok")
	}
}

func TestGetOne(t *testing.T) {
	cfg, err := config.Get("/etc/yascrapy/core.json")
	if err != nil {
		t.Error(err.Error())
	}
	node := config.SSDBNode{
		Host: cfg.ProxyRedisIp,
		Port: cfg.ProxyRedisPort,
	}
	client := ssdb.GetProxyClient(node, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	content, err := GetOne(client, "http_china")
	if err != nil {
		t.Error("TestGetOne fail ", err.Error())
	} else {
		t.Log(content)
		t.Log("TestGetOne ok")
	}
}

func TestDelete(t *testing.T) {
	cfg, err := config.Get("/etc/yascrapy/core.json")
	if err != nil {
		t.Error(err.Error())
	}
	node := config.SSDBNode{
		Host: cfg.ProxyRedisIp,
		Port: cfg.ProxyRedisPort,
	}
	client := ssdb.GetProxyClient(node, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	err = Delete(client, "127.0.0.1:4001", "http_china")
	if err != nil {
		t.Error("TestDelete fail ", err.Error())
	} else {
		t.Log("TestDelete ok")
	}
}
