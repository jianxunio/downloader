package core_queue

import "fmt"
import _ "math"
import "math/rand"
import "errors"
import "io/ioutil"
import "net/http"
import "downloader/config"
import "downloader/rabbitmq"
import "encoding/json"
import "strings"
import "github.com/streadway/amqp"

type Queue struct {
	Name  string `json:name`
	VHost string `json:vhost`
	Count int    `json:count`
}

func GetCrawler(q Queue, prefix string) string {
	s := strings.TrimPrefix(q.Name, prefix)
	return strings.Split(s, ":")[0]
}

func DeclareResponseQueues(conn *amqp.Connection, reqQueues []Queue, cnt int) int {
	ret := 0
	crawlers := make(map[string]int)
	for _, q := range reqQueues {
		c := GetCrawler(q, "http_request:")
		_, exists := crawlers[c]
		if !exists {
			crawlers[c] = 1
		}
	}
	for c, _ := range crawlers {
		for i := 0; i < cnt; i++ {
			respQueueName := fmt.Sprintf("http_response:%s:%d", c, i)
			err := rabbitmq.DeclareQueue(conn, respQueueName)
			if err == nil {
				ret += 1
			} else {
				fmt.Println(err)
			}
		}
	}
	return ret
}

func GetRandRespQueue(queues []Queue, crawler string) (string, error) {
	crawlerRespQueues := make([]Queue, 0)
	for _, q := range queues {
		qCrawler := GetCrawler(q, "http_response:")
		if qCrawler == crawler {
			crawlerRespQueues = append(crawlerRespQueues, q)
		}
	}
	cnt := len(crawlerRespQueues)
	if cnt <= 0 {
		return "", errors.New(fmt.Sprintf("respQueueCount %d", cnt))
	}
	index := rand.Intn(cnt)
	return crawlerRespQueues[index].Name, nil
}

func GetRandReqQueue(queues []Queue, crawler string) (string, error) {
	crawlerReqQueues := make([]Queue, 0)
	for _, q := range queues {
		qCrawler := GetCrawler(q, "http_request:")
		if qCrawler == crawler {
			crawlerReqQueues = append(crawlerReqQueues, q)
		}
	}
	cnt := len(crawlerReqQueues)
	if cnt <= 0 {
		return "", errors.New(fmt.Sprintf("reqQueueCount %d", cnt))
	}
	index := rand.Intn(cnt)
	return crawlerReqQueues[index].Name, nil
}

// Ensure that queues are not empty.
// Use QueueWeight as random choice weight.
// QueueWeight = (Log10(MAX_QUEUE_COUNT) / Log10(q.COUNT)) * 100.
// Need add this line with the main function `rand.Seed(time.Now().Unix())`
// func GetRandReqQueue(queues []Queue) (Queue, error) {
// 	totals := make([]int, 0)
// 	runningTotal := 0
// 	weight := 0
// 	max := 10
// 	for _, q := range queues {
// 		if q.Count > max {
// 			max = q.Count
// 		}
// 	}
// 	for _, q := range queues {
// 		if q.Count < 10 {
// 			q.Count = 10
// 		}
// 		weight = (int(math.Log10(float64(max))) * 100) / int(math.Log10(float64(q.Count)))
// 		runningTotal += weight
// 		totals = append(totals, runningTotal)
// 	}
// 	if runningTotal <= 0 {
// 		return Queue{}, errors.New(fmt.Sprintf("queues len %d, runningTotal %d", len(queues), runningTotal))
// 	}
// 	rnd := rand.Intn(runningTotal)
// 	for i, total := range totals {
// 		if rnd < total {
// 			return queues[i], nil
// 		}
// 	}
// 	return queues[0], nil
// }

func RemoveEmptyQueues(queues []Queue) []Queue {
	retQueues := make([]Queue, 0)
	for _, q := range queues {
		if q.Count > 0 {
			retQueues = append(retQueues, q)
		}
	}
	return retQueues
}

func GetQueues(prefix string, cfg config.Config) ([]Queue, error) {
	prefixQueues := make([]Queue, 0)
	queues := make([]Queue, 0)
	manager := fmt.Sprintf("http://%s:%d/api/queues/", cfg.RabbitmqManagerIp, cfg.RabbitmqManagerPort)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", manager, nil)
	req.SetBasicAuth(cfg.RabbitmqUser, cfg.RabbitmqPassword)
	resp, err := client.Do(req)
	if err != nil {
		return queues, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return queues, err
	}
	err = json.Unmarshal(body, &queues)
	if err != nil {
		return queues, err
	}
	for _, q := range queues {
		if strings.Contains(q.Name, prefix) {
			prefixQueues = append(prefixQueues, q)
		}
	}
	return prefixQueues, err
}

// Remove inaccessible queues and update queues count
func UpdateQueueCount(queues []Queue, conn *amqp.Connection) ([]Queue, error) {
	retQueues := make([]Queue, 0)
	for _, reqQueue := range queues {
		reqChannel, err := conn.Channel()
		if err != nil {
			return retQueues, err
		}
		q, err := reqChannel.QueueInspect(reqQueue.Name)
		if err != nil {
			reqChannel.Close()
			continue
		} else {
			reqQueue.Count = q.Messages
		}
		retQueues = append(retQueues, reqQueue)
		reqChannel.Close()
	}
	return retQueues, nil
}
