package http

import "testing"
import "downloader/request"
import "downloader/config"
import "downloader/ssdb"

func initR() *request.HttpRequest {
	headers := make(map[string]interface{})
	headers["referer"] = "http://meituan.com"
	params := make(map[string]interface{})
	params["foo"] = "bar"
	params["status"] = 1.1
	cookies := make(map[string]interface{})
	cookies["isLogin"] = 0
	r := &request.HttpRequest{
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

func TestSend(t *testing.T) {
	cfg, err := config.Get("/etc/yascrapy/core.json")
	if err != nil {
		t.Error(err.Error())
	}
	proxyNode := config.SSDBNode{
		Host: cfg.ProxyRedisIp,
		Port: cfg.ProxyRedisPort,
	}
	httpClient := GetClient()
	proxyClient := ssdb.GetProxyClient(proxyNode, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	r := initR()
	resp, err := Send(proxyClient, r, httpClient)
	// t.Log(resp.Html)
	t.Log(resp.Encoding)
	if err != nil {
		t.Error("TestSend error:", err.Error())
	} else {
		t.Log("TestSend success")
	}
}

func TestHttpsSend(t *testing.T) {
	cfg, err := config.Get("/etc/yascrapy/core.json")
	if err != nil {
		t.Error(err.Error())
	}
	proxyNode := config.SSDBNode{
		Host: cfg.ProxyRedisIp,
		Port: cfg.ProxyRedisPort,
	}
	httpClient := GetClient()
	proxyClient := ssdb.GetProxyClient(proxyNode, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	r := initR()
	r.Url = "https://github.com/cphilo"
	r.Params = make(map[string]interface{})
	r.ProxyName = "https_oversea"
	resp, err := Send(proxyClient, r, httpClient)
	t.Log(resp.HttpProxy)
	t.Log(resp.Encoding)
	t.Log(resp.Html)
	if err != nil {
		t.Error("TestSend error:", err.Error())
	} else {
		t.Log("TestSend success")
	}
}

func TestCookiesSend(t *testing.T) {
	cfg, err := config.Get("/etc/yascrapy/core.json")
	if err != nil {
		t.Error(err.Error())
	}
	proxyNode := config.SSDBNode{
		Host: cfg.ProxyRedisIp,
		Port: cfg.ProxyRedisPort,
	}
	httpClient := GetClient()
	proxyClient := ssdb.GetProxyClient(proxyNode, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	r := &request.HttpRequest{
		Url:         "http://m.weibo.cn/page/json?containerid=1005053122388424_-_FANS",
		Method:      "Get",
		Data:        "",
		Timeout:     15,
		ProxyName:   "http_china",
		CrawlerName: "test_crawler",
	}
	r.Headers = make(map[string]interface{})
	headers := map[string]string{
		"User-Agent":                "Mozilla/5.0 (iPhone; CPU iPhone OS 7_0 like Mac OS X; en-us) AppleWebKit/537.51.1 (KHTML, like Gecko) Version/7.0 Mobile/11A465 Safari/9537.53",
		"Accept-Language":           "zh-CN,zh;q=0.8,en;q=0.6",
		"Connection":                "keep-alive",
		"Origin":                    "https://passport.weibo.cn",
		"Accept":                    "*/*",
		"Upgrade-Insecure-Requests": "1",
	}
	for k, v := range headers {
		r.Headers[k] = v
	}
	cookies := make(map[string]interface{})
	cookies["SSOLoginState"] = "1457958715"
	cookies["SUB"] = "_2A2574t9rDeRxGeNG7FMY8SjJyzuIHXVZLOEjrDV6PUJbrdBeLWrEkW1LHetCDwRavAlEQLmOV3MeWp4y3LtbKg.."
	cookies["SUHB"] = "0ftps-U0SQ2EZE"
	r.Cookies = cookies
	resp, err := Send(proxyClient, r, httpClient)
	t.Log(resp.HttpProxy)
	t.Log(resp.Html)
	if err != nil {
		t.Error("TestCookiesSend fail", err.Error())
	} else {
		t.Log("TestCookiesSend success")
	}
}
