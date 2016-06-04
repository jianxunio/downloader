package request

import (
	"downloader/config"
	"downloader/rabbitmq"
	"fmt"
	"testing"
)

func initR() HttpRequest {
	headers := make(map[string]interface{})
	headers["referer"] = "http://meituan.com"
	params := make(map[string]interface{})
	params["foo"] = "bar"
	params["status"] = 1
	cookies := make(map[string]interface{})
	cookies["isLogin"] = 0
	r := HttpRequest{
		Method:      "GET",
		Url:         "http://wh.meituan.com",
		Headers:     headers,
		Data:        "",
		Params:      params,
		Cookies:     cookies,
		Timeout:     15,
		ProxyName:   "http_china",
		CrawlerName: "test_crawler",
	}
	return r
}

func TestPush(t *testing.T) {
	cfg, err := config.Get("/etc/yascrapy/core.json")
	queueName := "http_request:test_crawler"
	if err != nil {
		t.Error(err.Error())
	}
	conn, err := rabbitmq.GetConn(cfg)
	if err != nil {
		t.Error(err.Error())
	}
	err = rabbitmq.DeclareQueue(conn, queueName)
	if err != nil {
		t.Error(err.Error())
	}
	ch, err := rabbitmq.MakeChannel(conn)
	if err != nil {
		t.Error(err.Error())
	}
	defer conn.Close()
	defer ch.Close()
	r := initR()
	err = Push(ch, r, queueName)
	if err != nil {
		t.Error("TestPush fail ", err.Error())
	} else {
		t.Log("TestPush ok")
	}
}

/*
func TestGet(t *testing.T) {
    conn := rabbitmq.GetConn()
    ch, err := MakeChannel(conn,"test_crawler")
    defer ch.Close()
    defer conn.Close()
    if err != nil {
        t.Error("TestPush fail", err.Error())
    }
    r, err := Get(ch, "test_crawler")
    if err != nil {
        t.Error("TestGet fail", err.Error())
    } else {
        t.Log("TestGet success")
    }
    err = Done(ch, r, true)
    if err != nil {
        t.Error("TestDelete fail ", err.Error())
    } else {
        t.Log("TestDelete ok")
    }
}
*/

func BenchmarkPush(b *testing.B) {
	queueName := "http_request:test_crawler"
	cfg, err := config.Get("/etc/yascrapy/core.json")
	if err != nil {
		b.Error(err.Error())
	}
	conn, err := rabbitmq.GetConn(cfg)
	if err != nil {
		b.Error(err.Error())
	}
	err = rabbitmq.DeclareQueue(conn, queueName)
	if err != nil {
		b.Error(err.Error())
	}
	ch, err := rabbitmq.MakeChannel(conn)
	if err != nil {
		b.Log("Get Channel", err.Error())
	}
	for i := 0; i < b.N; i++ {
		r := initR()
		r.Url = fmt.Sprintf("%s%d", r.Url, i)
		err = Push(ch, r, queueName)
		if err != nil {
			b.Log(err.Error())
		}
	}
	ch.Close()
	conn.Close()
}

/*
func BenchmarkGet(b *testing.B) {
    cfg := config.Get()
    conn := rabbitmq.GetConn(cfg)
    rabbitmq.DeclareQueue(conn, GetQueueName("stackoverflow"))
    ch, err := rabbitmq.MakeChannel(conn)
    if err != nil {
        b.Log("Get Channel fail", err.Error())
    }
    for i := 0; i < b.N; i++ {
        r, err := Get(ch, "stackoverflow")
        if err != nil {
            b.Log(err.Error())
        }
        err = Done(ch, r, true)
        if err != nil {
            b.Log("TestDelete fail ", err.Error())
        }
    }
    ch.Close()
    conn.Close()
}*/
