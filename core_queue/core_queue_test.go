package core_queue

import "testing"
import "fmt"
import "math/rand"
import "time"
import "downloader/config"

func TestGetQueues(t *testing.T) {
	cfg, err := config.Get("/etc/yascrapy/core.json")
	if err != nil {
		t.Error(err.Error())
	}
	queues, err := GetQueues("", cfg)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(queues)
	}
	t.Error("to see output")
}

func TestGetRandQueue(t *testing.T) {
	var q Queue
	queues := make([]Queue, 0)
	queueCnts := []int{9, 200, 5000, 300, 600000}
	retCnt := []int{0, 0, 0, 0, 0}
	rand.Seed(time.Now().Unix())
	for i, cnt := range queueCnts {
		q = Queue{
			Name:  fmt.Sprintf("http_request:crawler%d", i),
			VHost: "/",
			Count: cnt,
		}
		queues = append(queues, q)
	}
	for i := 0; i < 100; i++ {
		q, err := GetRandReqQueue(queues)
		if err != nil {
			t.Error(err.Error())
		}
		for j := 0; j < len(queueCnts); j++ {
			if queueCnts[j] == q.Count {
				retCnt[j] += 1
			}
		}
	}
	t.Log(retCnt)
	t.Error("to see output")
}
