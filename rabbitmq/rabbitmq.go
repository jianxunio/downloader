package rabbitmq

import "downloader/config"
import "github.com/streadway/amqp"
import "fmt"

func CreateConn(cfg config.Config) (interface{}, error) {
	return GetConn(cfg)
}

func GetConn(cfg config.Config) (*amqp.Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d", cfg.RabbitmqUser, cfg.RabbitmqPassword, cfg.RabbitmqIp, cfg.RabbitmqPort)
	conn, err := amqp.Dial(url)
	if err != nil {
		return conn, err
	}
	return conn, nil
}

func ConnClosed(cfg config.Config) bool {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d", cfg.RabbitmqUser, cfg.RabbitmqPassword, cfg.RabbitmqIp, cfg.RabbitmqPort)
	conn, err := amqp.Dial(url)
	if err != nil {
		return true
	}
	_, err = conn.Channel()
	if err != nil {
		return true
	}
	conn.Close()
	return false
}

func DeclareQueue(conn *amqp.Connection, queueName string) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	args := make(amqp.Table, 0)
	args["x-max-length"] = int64(1000000)
	defer ch.Close()
	_, err = ch.QueueDeclare(queueName, true, false, false, false, args)
	if err != nil {
		return err
	}
	return nil
}

func MakeChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return ch, err
	}
	err = ch.Qos(1, 0, false)
	if err != nil {
		return ch, err
	}
	ch.Confirm(false)
	return ch, err
}
