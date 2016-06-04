package main

import "downloader/request"
import "downloader/config"
import "downloader/ssdb"
import local_http "downloader/http"
import "downloader/utils"
import "net/http"
import "fmt"
import "flag"
import "sync"

func initR(crawlerUrl string, crawlerName string, proxyName string, isMobile bool) *request.HttpRequest {
	headers := make(map[string]interface{})
	params := make(map[string]interface{})
	cookies := make(map[string]interface{})

	r := &request.HttpRequest{
		Method:      "GET",
		Url:         crawlerUrl,
		Headers:     headers,
		Data:        "",
		Params:      params,
		Cookies:     cookies,
		Timeout:     15,
		ProxyName:   proxyName,
		CrawlerName: fmt.Sprintf("%s__test", crawlerName),
	}
	if isMobile {
		newHeaders := map[string]string{
			"User-Agent":                "Mozilla/5.0 (iPhone; CPU iPhone OS 7_0 like Mac OS X; en-us) AppleWebKit/537.51.1 (KHTML, like Gecko) Version/7.0 Mobile/11A465 Safari/9537.53",
			"Accept-Language":           "zh-CN,zh;q=0.8,en;q=0.6",
			"Connection":                "keep-alive",
			"Accept":                    "*/*",
			"Upgrade-Insecure-Requests": "1",
			"Cache-Control":             "max-age=0",
			"Host":                      "m.weibo.cn",
		}
		for k, v := range newHeaders {
			r.Headers[k] = v
		}
	}
	return r
}

func sendRequest(client *ssdb.ProxyClient, r *request.HttpRequest, httpClient *http.Client) {
	for {
		resp, err := local_http.Send(client, r, httpClient)
		if err != nil {
			utils.Error.Println("http send fail ", err.Error())
		} else {
			if resp.StatusCode == 200 {
				utils.Info.Println("status 200, success")
			} else {
				utils.Warning.Println("status ", resp.StatusCode, resp.Reason)
			}
		}
	}
}

func main() {
	flag.Usage = func() {
		fmt.Println("This program is used to test sending request to url with http proxy")
		flag.PrintDefaults()
	}

	logLevel := flag.String("level", "info", "log level: debug, info, warning, error")
	logFile := flag.String("log_file", "/var/log/yascrapy/crawler.log", "specify the output log file")
	proxyName := flag.String("proxy_name", "http_china", "specify http proxy config")
	crawlerName := flag.String("crawler_name", "crawler", "specify crawler name like github, stackoverflow")
	crawlerUrl := flag.String("url", "https://github.com/cphilo", "specify crawler url")
	crawlerConsumers := flag.Int("crawler_consumers", 100, "specify crawler consumers, max 5000")
	isMobile := flag.Bool("mobile", false, "if http mobile header, set True")
	cfgFile := flag.String("config_file", "/etc/yascrapy/proxy.json", "specify proxy config file to get proxy redis ip and port")
	flag.Parse()
	handles, err := utils.GetHandles(*logLevel, *logFile)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	utils.InitLoggers(handles[0], handles[1], handles[2], handles[3])
	run(*proxyName, *crawlerName, *crawlerUrl, *crawlerConsumers, *isMobile, *cfgFile)
}

func run(proxyName string, crawlerName string, crawlerUrl string, crawlerConsumers int, isMobile bool, cfgFile string) {
	cfg, err := config.GetProxyConfig(cfgFile)
	if err != nil {
		utils.Error.Println(err.Error())
		return
	}
	proxyNode := config.SSDBNode{
		Host: cfg.ProxyRedisIp,
		Port: cfg.ProxyRedisPort,
	}
	proxyClient := ssdb.GetProxyClient(proxyNode, ssdb.Params{MaxIdle: 5000, MaxActive: 5000})
	r := initR(crawlerUrl, crawlerName, proxyName, isMobile)
	httpClient := local_http.GetClient()
	var wg sync.WaitGroup
	wg.Add(1)
	for i := 0; i < crawlerConsumers; i++ {
		go sendRequest(proxyClient, r, httpClient)
	}
	wg.Wait()
}
