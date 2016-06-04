package core_consumer

import "testing"
import "downloader/config"
import "downloader/core_queue"
import "downloader/rabbitmq"
import "github.com/streadway/amqp"
import "math/rand"
import "time"

func TestCoreConsumer(t *testing.T) {
	cfg, err := config.Get("/etc/yascrapy/core.json")
	if err != nil {
		t.Error(err.Error())
	}
	var consumers []*CoreConsumer
	var conn *amqp.Connection
	conn, err = rabbitmq.GetConn(cfg)
	if err != nil {
		t.Error("rabbitmq GetConn fail", err.Error())
	}
	queues, err := core_queue.GetQueues("http_request:", cfg)
	t.Log("queues cnt", len(queues))
	if err != nil {
		t.Error("Get Queues fail", err.Error())
	}
	queues, err = core_queue.UpdateQueueCount(queues, conn)
	if err != nil {
		t.Error("UpdateQueueCount fail", err.Error())
	}
	queues = core_queue.RemoveEmptyQueues(queues)
	t.Log("queues cnt", len(queues))
	rand.Seed(time.Now().Unix())
	consumers = AliveCoreConsumers(consumers)
	respQueues, err := core_queue.GetQueues("http_response:", cfg)
	if err != nil {
		t.Error("http response error", err.Error())
	}
	for i := 0; i < cfg.RabbitmqConsumers; i++ {
		reqQueue, err := core_queue.GetRandReqQueue(queues)
		if err != nil {
			t.Error(err.Error())
		}
		crawler := core_queue.GetCrawler(reqQueue, "http_request:")
		respQueueName, err := core_queue.GetRandRespQueue(respQueues, crawler)
		aliveReqConsumers := AliveReqConsumers(reqQueue.Name, consumers)
		aliveRespConsumers := AliveRespConsumers(respQueueName, consumers)
		if err != nil {
			t.Error(err.Error())
		}
		t.Log("req queue", reqQueue.Name, "resp queue", respQueueName)
		t.Log("aliveReqConsumers ", len(aliveReqConsumers), "aliveRespConsumers", len(aliveRespConsumers))
		c := CreateCoreConsumer(i, crawler, reqQueue.Name, respQueueName)
		consumers = append(consumers, c)
	}
	t.Error("to see output")
}
