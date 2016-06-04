package response

import (
	"downloader/config"
	"downloader/rabbitmq"
	"downloader/request"
	"downloader/ssdb"
	"fmt"
	"testing"
)

func initRequest() request.HttpRequest {
	headers := make(map[string]interface{})
	headers["referer"] = "http://meituan.com"
	params := make(map[string]interface{})
	params["foo"] = "bar"
	params["status"] = 1
	cookies := make(map[string]interface{})
	cookies["isLogin"] = 0
	r := request.HttpRequest{
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

func initResponse(req request.HttpRequest) HttpResponse {
	resp := HttpResponse{
		Request:     req,
		ErrorCode:   0,
		ErrorMsg:    "",
		StatusCode:  200,
		Reason:      "",
		Html:        "<p>Hello, world!</p>",
		Url:         req.Url,
		Headers:     req.Headers,
		Cookies:     req.Cookies,
		Encoding:    "utf-8",
		CrawlerName: req.CrawlerName,
		ProxyName:   "http_china",
		HttpProxy:   "",
	}
	return resp
}

func TestPush(t *testing.T) {
	cfg, err := config.Get()
	if err != nil {
		t.Error(err.Error())
	}
	nodes := cfg.SSDBNodes
	clients := ssdb.GetClients(nodes, cfg, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	req := initRequest()
	r := initResponse(req)
	err = Push(clients, r)
	if err != nil {
		t.Error("TestPush fail ", err.Error())
	} else {
		t.Log("TestPush ok")
	}
}

func TestPushKey(t *testing.T) {
	queueName := "test_crawler"
	req := initRequest()
	r := initResponse(req)
	cfg, err := config.Get("/etc/yascrapy/core.json")
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
	defer ch.Close()
	if err != nil {
		t.Error("TestPushKey fail %s", err.Error())
	}
	err = PushKey(ch, r, queueName)
	if err != nil {
		t.Error("TestPushKey fail ", err.Error())
	} else {
		t.Log("TestPushKey ok")
	}
}

func BenchmarkPushKey(b *testing.B) {
	queueName := "test_crawler"
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
	defer ch.Close()
	if err != nil {
		b.Error("TestPush fail", err.Error())
	}
	for i := 0; i < b.N; i++ {
		req := initRequest()
		req.Url = fmt.Sprintf("%s%d", req.Url, i)
		resp := initResponse(req)
		err = PushKey(ch, resp, queueName)
		if err != nil {
			b.Log(err.Error())
		}
	}
}

func BenchmarkPush(b *testing.B) {
	cfg, err := config.Get("/etc/yascrapy/core.json")
	if err != nil {
		b.Error(err.Error())
	}
	nodes := cfg.SSDBNodes
	clients := ssdb.GetClients(nodes, cfg, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	for i := 0; i < b.N; i++ {
		req := initRequest()
		req.Url = fmt.Sprintf("%s%d", req.Url, i)
		resp := initResponse(req)
		err := Push(clients, resp)
		if err != nil {
			b.Log("TestPush fail ", err.Error())
		}
	}
}
