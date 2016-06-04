package ssdb

import "testing"
import "downloader/config"

func TestGetClient(t *testing.T) {
	var k string
	var client *SSDBClient
	var err error
	cfg, err := config.Get("/etc/yascrapy/core.json")
	if err != nil {
		t.Error(err.Error())
	}
	nodes := cfg.SSDBNodes
	clients := GetClients(nodes, cfg, Params{MaxActive: 100, MaxIdle: 100})
	k = "http_response:stackoverflow:dkfjkd"
	client, err = GetClient(clients, k)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(client.Node.Host, client.Node.Port)
	k = "http_response:test_crawler:djfdjk"
	client, err = GetClient(clients, k)
	t.Log(client.Node.Host, client.Node.Port)
	k = "http_response:test:djfdjk"
	client, err = GetClient(clients, k)
	t.Log(client.Node.Host, client.Node.Port)
	k = "http_response:test348939:djfdjk"
	client, err = GetClient(clients, k)
	t.Log(client.Node.Host, client.Node.Port)
}
