package rabbitmq

import "testing"
import "reflect"
import "github.com/streadway/amqp"
import "downloader/config"

func TestConnPool(t *testing.T) {
	cfg, err := config.Get()
	if err != nil {
		t.Error("get config error", err.Error())
	}
	p := &ConnPool{
		MaxIdle:   10,
		MaxActive: 10,
		Dial:      CreateConn,
		Cfg:       cfg,
	}
	var c *amqp.Connection
	for i := 0; i < p.MaxActive; i++ {
		c = p.Get().(*amqp.Connection)
		t.Log(reflect.TypeOf(c))
		// p.Release(c)
	}
	c = p.Get().(*amqp.Connection)
	t.Error("see output")
}
