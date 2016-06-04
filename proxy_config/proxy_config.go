package proxy_config

import "encoding/json"
import "downloader/config"
import "downloader/ssdb"
import "github.com/garyburd/redigo/redis"

func GetProxyNames(client *ssdb.ProxyClient) ([]string, error) {
	rc := client.ConnPool
	db := rc.Get()
	defer db.Close()
	return redis.Strings(db.Do("hkeys", redis.Args{}.Add("proxy_config")...))
}

func UpdateStatus(client *ssdb.ProxyClient, proxyName string, status int) error {
	cfg, err := Get(client, proxyName)
	if err != nil {
		return err
	}
	cfg.Status = status
	err = Add(client, cfg)
	return err
}

func Get(client *ssdb.ProxyClient, proxyName string) (config.ProxyConfigNode, error) {
	var h config.ProxyConfigNode
	rc := client.ConnPool
	db := rc.Get()
	defer db.Close()
	content, err := redis.String(db.Do("hget", redis.Args{}.Add("proxy_config").AddFlat(proxyName)...))
	if err != nil {
		return h, err
	}
	err = json.Unmarshal([]byte(content), &h)
	return h, err
}

func Add(client *ssdb.ProxyClient, p config.ProxyConfigNode) error {
	rc := client.ConnPool
	content, err := json.Marshal(p)
	if err != nil {
		return err
	}
	db := rc.Get()
	defer db.Close()
	_, err = db.Do("hset", redis.Args{}.Add("proxy_config").AddFlat(p.Name).AddFlat(string(content))...)
	return err
}

func Delete(client *ssdb.ProxyClient, proxyName string) error {
	rc := client.ConnPool
	db := rc.Get()
	defer db.Close()
	_, err := db.Do("hdel", redis.Args{}.Add("proxy_config").AddFlat(proxyName)...)
	return err
}
