package proxy

import "fmt"
import "github.com/garyburd/redigo/redis"
import "downloader/ssdb"
import "errors"
import "regexp"

func Push(client *ssdb.ProxyClient, proxy string, ProxyName string) error {
	ok, err := regexp.MatchString("^([0-9.]*:[0-9]*)", proxy)
	if err != nil {
		return err
	}
	if ok == false {
		return errors.New(fmt.Sprintf("invalid proxy %s", proxy))
	}
	k := fmt.Sprintf("http_proxy:%s", ProxyName)
	rc := client.ConnPool
	db := rc.Get()
	defer db.Close()
	_, err = db.Do("sadd", redis.Args{}.AddFlat(k).AddFlat(proxy)...)
	return err
}

func Get(client *ssdb.ProxyClient, ProxyName string) ([]string, error) {
	k := fmt.Sprintf("http_proxy:%s", ProxyName)
	rc := client.ConnPool
	db := rc.Get()
	defer db.Close()
	return redis.Strings(db.Do("smembers", redis.Args{}.AddFlat(k)...))
}

func GetCount(client *ssdb.ProxyClient, ProxyName string) (int, error) {
	k := fmt.Sprintf("http_proxy:%s", ProxyName)
	rc := client.ConnPool
	db := rc.Get()
	defer db.Close()
	return redis.Int(db.Do("scard", redis.Args{}.AddFlat(k)...))
}

func GetOne(client *ssdb.ProxyClient, ProxyName string) (string, error) {
	k := fmt.Sprintf("http_proxy:%s", ProxyName)
	rc := client.ConnPool
	db := rc.Get()
	defer db.Close()
	return redis.String(db.Do("srandmember", redis.Args{}.AddFlat(k)...))
}

func Delete(client *ssdb.ProxyClient, proxy string, ProxyName string) error {
	k := fmt.Sprintf("http_proxy:%s", ProxyName)
	rc := client.ConnPool
	db := rc.Get()
	defer db.Close()
    _, err := db.Do("srem", redis.Args{}.AddFlat(k).AddFlat(proxy)...)
	return err
}
