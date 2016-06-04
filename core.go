package main

import "time"
import "downloader/config"
import "downloader/response"
import "downloader/request"
import local_http "downloader/http"
import "net/http"
import "downloader/utils"
import "github.com/streadway/amqp"
import "downloader/ssdb"
import "downloader/rabbitmq"
import "downloader/core_consumer"
import "downloader/core_queue"
import "downloader/proxy"
import "flag"
import "fmt"
import "os"
import "os/signal"
import "syscall"
import "math/rand"
import "strings"

func CrawlerHttpClients(consumers int, maxConn int) []*http.Client {
	httpClients := make([]*http.Client, 0)
	totalClients := consumers / maxConn
	if consumers%maxConn > 0 {
		totalClients += 1
	}
	for i := 0; i < totalClients; i++ {
		httpClients = append(httpClients, local_http.GetClient())
	}
	return httpClients
}

// Tranverse proxy string to string formatted as "Host:Port".
func pureProxy(s string) string {
	if strings.HasPrefix(s, "https://") {
		return strings.Replace(s, "https://", "", 1)
	} else if strings.HasPrefix(s, "http://") {
		return strings.Replace(s, "http://", "", 1)
	} else {
		return s
	}
}

func consume(httpClient *http.Client, proxyClient *ssdb.ProxyClient, req *request.HttpRequest,
	ssdbClients []*ssdb.SSDBClient, reqChannel *amqp.Channel,
	respChannel *amqp.Channel, crawlerName string, respQueueName string) error {
	resp, err := local_http.Send(proxyClient, req, httpClient)
	if err != nil {
		utils.Info.Println("http Send fail", crawlerName, err.Error(), resp.HttpProxy)

		// Remove http_proxy in redis to improve proxy quality.
		if resp.HttpProxy != "" {
			p := pureProxy(resp.HttpProxy)
			utils.Info.Println("delete proxy ", p)
			err = proxy.Delete(proxyClient, p, req.ProxyName)
			if err != nil {
				utils.Error.Println("delete proxy fail", err.Error())
			}
		}

		return request.Done(reqChannel, req, false)
	}
	err = response.Push(ssdbClients, resp)
	if err != nil {
		utils.Error.Println("response Push fail", crawlerName, err.Error())
		return request.Done(reqChannel, req, false)
	}
	err = response.PushKey(respChannel, resp, respQueueName)
	if err != nil {
		utils.Error.Println("response PushKey fail", crawlerName, err.Error())
		return request.Done(reqChannel, req, false)
	}
	return request.Done(reqChannel, req, true)
}

func consumer(
	c *core_consumer.CoreConsumer,
	connPool *rabbitmq.ConnPool,
	ssdbClients []*ssdb.SSDBClient,
	httpClient *http.Client,
	proxyClient *ssdb.ProxyClient) {
	defer core_consumer.ExitCoreConsumer(c)
	utils.Info.Printf("[%s] start consumer", c.CrawlerName)
	conn := connPool.Get().(*amqp.Connection)
	connPool.Release(conn)
	reqChannel, err := conn.Channel()
	if err != nil {
		utils.Error.Println("request MakeChannel fail and exit ", c.CrawlerName, err.Error())
		return
	} else {
		utils.Info.Println(c.CrawlerName, "request MakeChannel ok")
	}
	defer reqChannel.Close()
	err = reqChannel.Qos(30, 0, false)
	if err != nil {
		utils.Error.Println("request channel Qos fail and exit", c.CrawlerName, err.Error())
		return
	}
	respChannel, err := conn.Channel()
	respChannel.Confirm(false)
	if err != nil {
		utils.Error.Println("response MakeChannel fail and exit ", c.CrawlerName, err.Error())
		return
	} else {
		utils.Info.Println(c.CrawlerName, "response MakeChannel ok")
	}
	defer respChannel.Close()
	deliveries, err := request.Consume(reqChannel, c.ReqQueue)
	if err != nil {
		utils.Error.Println("request Consume fail: ", c)
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			utils.Debug.Println("check channel status", c.CrawlerName, c.ReqQueue, c.RespQueue)
			q, err := reqChannel.QueueInspect(c.ReqQueue)
			if err != nil {
				utils.Info.Println(c.CrawlerName, "QueueInspect fail", err.Error())
				return
			} else if err == nil && q.Messages == 0 {
				utils.Info.Println("QueueInspect empty and exit")
				return
			} else {
				utils.Debug.Println("QueueInspect success", c)
			}
			q, err = respChannel.QueueInspect(c.RespQueue)
			if err != nil {
				utils.Info.Println(c.CrawlerName, "RespQueue QueueInspect fail", err.Error())
				return
			} else {
				utils.Debug.Println("QueueInspect success", c)
			}
		case d := <-deliveries:
			req, err := request.Get(d)
			if err == nil && req == nil {
				utils.Info.Printf("[%s]consumer Get Delivery empty\n", c.CrawlerName)
				_, err = reqChannel.QueueInspect(c.ReqQueue)
				if err != nil {
					utils.Info.Println(c.CrawlerName, "QueueInspect fail", err.Error())
					return
				} else {
					time.Sleep(1 * time.Second)
				}
				continue
			}
			if err != nil {
				utils.Info.Printf("[%s]request Get deliveries error: %s\n", c.CrawlerName, string(d.Body))
			} else {
				utils.Debug.Println("consume ", req.Url)
				go consume(httpClient, proxyClient, req, ssdbClients, reqChannel, respChannel, c.CrawlerName, c.RespQueue)
			}
		default:
			utils.Debug.Println("coming to default operation", c)
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func cleanup(cfg config.Config, proxyClient *ssdb.ProxyClient) {
	proxyClient.ConnPool.Close()
	utils.Info.Printf("process exit")
}

func run(cfgFile string, consumerNumber int, crawlerName string) {
	var conn *amqp.Connection
	var consumers []*core_consumer.CoreConsumer
	cfg, err := config.Get(cfgFile)
	if err != nil {
		utils.Error.Println("read config file fail ", err.Error(), cfgFile)
		return
	}
	proxyNode := config.SSDBNode{
		Host: cfg.ProxyRedisIp,
		Port: cfg.ProxyRedisPort,
	}
	ssdbNodes := cfg.SSDBNodes
	if len(ssdbNodes) == 0 {
		utils.Error.Println("ssdb nodes can not be empty")
		return
	}
	ssdbClients := ssdb.GetClients(ssdbNodes, cfg, ssdb.Params{MaxIdle: 5000, MaxActive: 5000})
	proxyClient := ssdb.GetProxyClient(proxyNode, ssdb.Params{MaxIdle: 5000, MaxActive: 5000})
	connPool := &rabbitmq.ConnPool{
		MaxIdle:   10,
		MaxActive: 50,
		Dial:      rabbitmq.CreateConn,
		Cfg:       cfg,
	}
	maxConn := 1000
	httpClients := CrawlerHttpClients(consumerNumber, maxConn)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	signal.Notify(sigs, syscall.SIGTERM)
	go func() {
		<-sigs
		cleanup(cfg, proxyClient)
		os.Exit(0)
	}()

	for {
		reqQueues, err := core_queue.GetQueues("http_request:", cfg)
		if err != nil {
			utils.Error.Printf("GetReqQueues error:%s\n", err.Error())
			time.Sleep(1 * time.Second)
			continue
		} else {
			utils.Debug.Println("GetReqQueues success")
		}

		respQueues, err := core_queue.GetQueues("http_response:", cfg)
		if err != nil {
			utils.Error.Printf("GetRespQueues error:%s\n", err.Error())
			time.Sleep(1 * time.Second)
			continue
		} else {
			utils.Debug.Println("GetRespQueues success")
		}

		conn = connPool.Get().(*amqp.Connection)
		reqQueues, err = core_queue.UpdateQueueCount(reqQueues, conn)
		if err != nil {
			utils.Error.Println(err.Error())
			time.Sleep(1 * time.Second)
			continue
		} else {
			utils.Debug.Println("UpdateQueueCount success")
		}
		connPool.Release(conn)

		reqQueues = core_queue.RemoveEmptyQueues(reqQueues)
		utils.Debug.Println(reqQueues)
		if len(reqQueues) == 0 {
			utils.Info.Println("req queues empty")
			time.Sleep(1 * time.Second)
			continue
		} else {
			utils.Debug.Println("RemoveEmptyQueues success")
		}
		core_queue.DeclareResponseQueues(conn, reqQueues, 5)

		conn = connPool.Get().(*amqp.Connection)
		respQueues, err = core_queue.UpdateQueueCount(respQueues, conn)
		utils.Debug.Println(respQueues)
		if err != nil {
			conn = nil
			utils.Error.Println(err.Error())
			time.Sleep(1 * time.Second)
			continue
		}
		connPool.Release(conn)

		if len(respQueues) == 0 {
			utils.Info.Println("resp queues empty")
			time.Sleep(1 * time.Second)
			continue
		}

		consumers = core_consumer.AliveCrawlerConsumers(consumers, crawlerName)
		utils.Debug.Println("total consumers: ", len(consumers), ", need start consumers:", consumerNumber-len(consumers))
		newConsumers := consumerNumber - len(consumers)

		for i := 0; i < newConsumers; i++ {
			reqQueueName, err := core_queue.GetRandReqQueue(reqQueues, crawlerName)
			if err != nil {
				utils.Error.Printf(err.Error())
				continue
			}
			utils.Debug.Printf("GetRandReqQueue %s\n", reqQueueName)
			respQueueName, err := core_queue.GetRandRespQueue(respQueues, crawlerName)
			if err != nil {
				utils.Error.Printf(err.Error())
				continue
			}
			utils.Debug.Printf("GetRandRespQueue %s\n", respQueueName)
			index := i / maxConn
			c := core_consumer.CreateCoreConsumer(i, crawlerName, reqQueueName, respQueueName)
			consumers = append(consumers, c)
			utils.Debug.Println("add consumer", c)
			go consumer(c, connPool, ssdbClients, httpClients[index], proxyClient)
		}
		if len(consumers) < consumerNumber {
			utils.Warning.Printf("aliveConsumers now %d, lower than %d\n", len(consumers), consumerNumber)
		}
		utils.Debug.Println("total consumers ", len(consumers))
		time.Sleep(1 * time.Second)
	}
}

func main() {
	logLevel := flag.String("level", "info", "log level: debug, info, warning, error")
	// logFile := flag.String("log_file", "/var/log/yascrapy/core.log", "specify the output log file")
	cfgFile := flag.String("config_file", "/etc/yascrapy/core.json", "specify core downloader config file in json")
	consumers := flag.Int("consumers", 1, "specify consumers number")
    crawlerName := flag.String("crawler_name", "", "specify crawler name, can not be empty")
	flag.Parse()
	// handles, err := utils.GetHandles(*logLevel, *logFile)
	handles, err := utils.GetHandles(*logLevel)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if *crawlerName == "" {
		fmt.Println("crawlerName can not be empty")
		return
	}
	utils.InitLoggers(handles[0], handles[1], handles[2], handles[3])
	rand.Seed(time.Now().Unix())
	run(*cfgFile, *consumers, *crawlerName)
}
