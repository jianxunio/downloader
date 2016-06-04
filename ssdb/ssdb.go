package ssdb

import "downloader/config"
import "fmt"
import "github.com/garyburd/redigo/redis"
import "time"
import "github.com/serialx/hashring"
import "errors"

type Params struct {
    MaxIdle int
    MaxActive int
}

type SSDBClient struct {
    Node config.SSDBNode
    ConnPool *redis.Pool
    Ring *hashring.HashRing
}

type ProxyClient struct {
    Node config.SSDBNode
    ConnPool *redis.Pool
}

func GetProxyClient(node config.SSDBNode, p Params) *ProxyClient {
    rc := GetPool(node, p)
    return &ProxyClient{Node: node,ConnPool: rc}
}

func GetClient(clients []*SSDBClient, key string)(*SSDBClient, error) {
    var ring *hashring.HashRing
    var server string
    if len(clients) > 0 {
        ring = clients[0].Ring
    } else {
        return nil, errors.New("ring.GetNode fail, clients empty")
    }
    foundServer, ok := ring.GetNode(key)
    if ok {
        for _, client := range clients {
            server = fmt.Sprintf("%s:%d", client.Node.Host, client.Node.Port)
            if server == foundServer {
                return client,nil
            }
        }
        return nil,errors.New("ring.GetNode ok, fail get client")
    } else {
        return nil,errors.New("ring.GetNode fail, clients empty")
    }
}

func GetClients(nodes []config.SSDBNode, cfg config.Config, p Params) []*SSDBClient {
    var server string
    servers := make([]string, 0)
    clients := make([]*SSDBClient, 0)
    for _, node := range nodes {
        rc := GetPool(node, p)
        clients = append(clients, &SSDBClient{Node: node,ConnPool: rc,Ring:nil,})
        server = fmt.Sprintf("%s:%d", node.Host, node.Port)
        servers = append(servers, server)
    }
    ring := hashring.New(servers)
    for _, client := range clients {
        client.Ring = ring
    }
    return clients
}

func GetPool(node config.SSDBNode, p Params) *redis.Pool {
    url := fmt.Sprintf("redis://%s:%d", node.Host, node.Port)
    if p.MaxIdle == 0 {
        p.MaxIdle = 10
    }
    if p.MaxActive == 0 {
        p.MaxActive = 10
    }
    rc := &redis.Pool{
        MaxIdle:     p.MaxIdle,
        MaxActive:   p.MaxActive,
        IdleTimeout: 180 * time.Second,
        Dial: func() (redis.Conn, error) {
            c, err := redis.DialURL(url)
            if err != nil {
                return nil, err
            }
            return c, nil
        },
    }
    return rc
}
