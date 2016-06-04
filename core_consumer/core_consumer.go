package core_consumer

type CoreConsumer struct {
	Tag         int
	CrawlerName string
	Status      int
	ReqQueue    string
	RespQueue   string
}

func CreateCoreConsumer(tag int, crawlerName string, reqQueue string, respQueue string) *CoreConsumer {
	return &CoreConsumer{
		Tag:         tag,
		CrawlerName: crawlerName,
		Status:      0,
		ReqQueue:    reqQueue,
		RespQueue:   respQueue,
	}
}

func AliveRespConsumers(respQueue string, consumers []*CoreConsumer) []*CoreConsumer {
	var aliveConsumers []*CoreConsumer
	for _, c := range consumers {
		if c.RespQueue == respQueue && c.Status == 0 {
			aliveConsumers = append(aliveConsumers, c)
		}
	}
	return aliveConsumers
}

func AliveReqConsumers(reqQueue string, consumers []*CoreConsumer) []*CoreConsumer {
	var aliveConsumers []*CoreConsumer
	for _, c := range consumers {
		if c.ReqQueue == reqQueue && c.Status == 0 {
			aliveConsumers = append(aliveConsumers, c)
		}
	}
	return aliveConsumers
}

func AliveCoreConsumers(consumers []*CoreConsumer) []*CoreConsumer {
	var newConsumers []*CoreConsumer
	for _, c := range consumers {
		if c.Status == 0 {
			newConsumers = append(newConsumers, c)
		}
	}
	return newConsumers
}

func AliveCrawlerConsumers(consumers []*CoreConsumer, crawler string) []*CoreConsumer {
	var newConsumers []*CoreConsumer
	for _, c := range consumers {
		if c.Status == 0 && c.CrawlerName == crawler {
			newConsumers = append(newConsumers, c)
		}
	}
	return newConsumers
}

func ExitCoreConsumer(c *CoreConsumer) {
	c.Status = 1
}
