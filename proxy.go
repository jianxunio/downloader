package main

import "fmt"
import "net/http"
import "net/url"
import "time"
import "io/ioutil"
import "strings"
import "downloader/config"
import "downloader/proxy"
import "downloader/proxy_config"
import "downloader/ssdb"
import "downloader/utils"
import "flag"
import "os"
import "os/signal"
import "syscall"

func checkProxy(proxyClient *ssdb.ProxyClient, h config.ProxyConfigNode, jobs chan string, results chan int) {
	for p := range jobs {
		u := fmt.Sprintf("http://%s", p)
		proxyUrl, err := url.Parse(u)
		if err != nil {
			utils.Info.Printf("parse url error: %s\n", err.Error())
		} else {
			client := http.Client{
				Timeout: time.Duration(time.Duration(h.CheckTimeout) * time.Second),
				Transport: &http.Transport{
					Proxy:               http.ProxyURL(proxyUrl),
					DisableKeepAlives:   true,
					TLSHandshakeTimeout: 5 * time.Second,
				},
			}
			resp, err := client.Get(h.CheckUrl)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == 200 {
					utils.Info.Printf("check url %s success\n", h.CheckUrl)
				} else {
					err := proxy.Delete(proxyClient, p, h.Name)
					if err != nil {
						utils.Error.Println("proxy Delete error: ", err.Error())
					}
				}
			} else {
				utils.Error.Printf("get url %s error: %s\n", h.CheckUrl, err.Error())
				err := proxy.Delete(proxyClient, p, h.Name)
				if err != nil {
					utils.Error.Println("proxy Delete error: ", err.Error())
				}
			}
		}
		results <- 0
	}
}

func CheckProxyWorker(proxyClient *ssdb.ProxyClient, h config.ProxyConfigNode, done chan bool) {
	for {
		select {
		case quit := <-done:
			if quit {
				utils.Warning.Printf("exit channel %s\n", h.Name)
				close(done)
				return
			}
		default:
			proxy_list, err := proxy.Get(proxyClient, h.Name)
			if err != nil {
				utils.Error.Printf("get %s proxy_list error:%s\n", h.Name, err.Error())
			} else {
				jobs := make(chan string, len(proxy_list))
				results := make(chan int, len(proxy_list))
				for _, proxy := range proxy_list {
					jobs <- proxy
				}
				close(jobs)
				for i := 0; i < h.CheckRequests; i++ {
					go checkProxy(proxyClient, h, jobs, results)
				}
				for i := 0; i < len(proxy_list); i++ {
					<-results
				}
				close(results)
			}
			time.Sleep(time.Duration(h.CheckFrequency) * time.Second)
		}
	}
}

func AddProxyWorker(proxyClient *ssdb.ProxyClient, h config.ProxyConfigNode, done chan bool) {
	client := http.Client{
		Timeout: time.Duration(time.Duration(h.AddTimeout) * time.Second),
	}
	for {
		select {
		case quit := <-done:
			if quit {
				utils.Warning.Printf("exit channel %s\n", h.Name)
				close(done)
				return
			}
		default:
			num, err := proxy.GetCount(proxyClient, h.Name)
			if err != nil {
				utils.Error.Println("proxy.GetCount fail")
			} else if num >= h.ProxyMax && h.ProxyMax > 0 {
				utils.Info.Printf("proxy %s full, %d items, max %d\n", h.Name, num, h.ProxyMax)
			} else {
				utils.Info.Printf("add proxy worker: %s\n", h.Name)
				resp, err := client.Get(h.ProxyUrl)
				if err == nil {
					content, err := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					if err != nil {
                        utils.Error.Printf("repsonse body to string error: %s \n", err.Error())
					} else {
						data := strings.Split(string(content), "\n")
						if len(data) < 10 {
							utils.Error.Println(string(content))
						}
						utils.Info.Printf("add %d proxies to queue\n", len(data))
						for _, d := range data {
							err := proxy.Push(proxyClient, d, h.Name)
							if err != nil {
								utils.Error.Printf("add proxy %s to %s queue error:%s\n", d, h.Name, err.Error())
							} else {
								utils.Debug.Printf("add proxy %s to queue\n", d)
							}
						}
					}
				}
			}
			time.Sleep(time.Duration(h.AddFrequency) * time.Second)
		}
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func clearChannels(cfgs []string, channels map[string]chan bool) {
	for name, channel := range channels {
		if contains(cfgs, name) == false {
			channel <- true
		}
	}
}

func deleteChannel(cfg string, channels map[string]chan bool) {
	if _, ok := channels[cfg]; ok {
		channels[cfg] <- true
	}
}

// `config.ProxyConfig` have `Status` field to avoid running proxy repeatedly.
// But we should set `Status` to `StatusStopped` when this process exit.
// Otherwise, you have to stop `config.ProxyConfig` manually.
func cleanup(client *ssdb.ProxyClient) {
	proxyNames, err := proxy_config.GetProxyNames(client)
	if err != nil {
		utils.Error.Printf("GetProxyNames error: %s\n", err.Error())
	}
	for _, name := range proxyNames {
		cfg, err := proxy_config.Get(client, name)
		if err != nil {
			utils.Error.Printf("proxy_config Get error: %s\n", err.Error())
		} else {
			proxy_config.UpdateStatus(client, cfg.Name, config.StatusStopped)
		}
	}
	client.ConnPool.Close()
}

func main() {
	logLevel := flag.String("level", "info", "log level: debug, info, warning, error")
	logFile := flag.String("log_file", "/var/log/yascrapy/proxy.log", "specify output log file")
	cfgFile := flag.String("config_file", "/etc/yascrapy/proxy.json", "specify yascrapy-proxy service config file")
	flag.Parse()
	handles, err := utils.GetHandles(*logLevel, *logFile)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	utils.InitLoggers(handles[0], handles[1], handles[2], handles[3])
	run(*cfgFile)
}

func run(cfgFile string) {
	cfg, err := config.GetProxyConfig(cfgFile)
	if err != nil {
		utils.Error.Println("GetProxyConfig fail", err.Error())
		return
	}
	proxyNode := config.SSDBNode{
		Host: cfg.ProxyRedisIp,
		Port: cfg.ProxyRedisPort,
	}
	proxyClient := ssdb.GetProxyClient(proxyNode, ssdb.Params{MaxActive: 100, MaxIdle: 100})
	AddProxyChannels := map[string]chan bool{}
	CheckProxyChannels := map[string]chan bool{}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		cleanup(proxyClient)
		os.Exit(0)
	}()

	for {
		proxyCfgs := make([]config.ProxyConfigNode, 0)
		proxyNames, err := proxy_config.GetProxyNames(proxyClient)
		if err != nil {
			utils.Error.Printf("GetProxyNames error: %s\n", err.Error())
		}

		for _, name := range proxyNames {
			cfg, err := proxy_config.Get(proxyClient, name)
			if err != nil {
				utils.Error.Printf("proxy_config Get error: %s\n", err.Error())
			} else {
				proxyCfgs = append(proxyCfgs, cfg)
			}
		}

		clearChannels(proxyNames, AddProxyChannels)
		clearChannels(proxyNames, CheckProxyChannels)

		for _, proxyCfg := range proxyCfgs {
			if proxyCfg.Status == config.StatusStopped {

				deleteChannel(proxyCfg.Name, AddProxyChannels)
				deleteChannel(proxyCfg.Name, CheckProxyChannels)

				addChannel := make(chan bool, 1)
				addChannel <- false
				AddProxyChannels[proxyCfg.Name] = addChannel

				checkChannel := make(chan bool, 1)
				checkChannel <- false
				CheckProxyChannels[proxyCfg.Name] = checkChannel

				go AddProxyWorker(proxyClient, proxyCfg, addChannel)
				go CheckProxyWorker(proxyClient, proxyCfg, checkChannel)

				err := proxy_config.UpdateStatus(proxyClient, proxyCfg.Name, config.StatusRunning)

				if err != nil {
					utils.Error.Printf("UpdateStatus %s error:%s\n", proxyCfg.Name, err.Error())
				}

			} else if proxyCfg.Status == config.StatusRunning {
				utils.Warning.Printf("%s StatusRunning\n", proxyCfg.Name)
			} else {
				utils.Error.Printf("status error:%s\n", err.Error())
			}
		}

		time.Sleep(180 * time.Second)
	}
}
