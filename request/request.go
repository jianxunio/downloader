package request

import "encoding/json"
import "github.com/streadway/amqp"
import "errors"

type HttpRequest struct {
    Method string
    Url string
    Headers map[string] interface{}
    Data string
    Params map[string] interface{}
    Cookies map[string] interface{}
    Timeout int
    ProxyName string
    CrawlerName string
    DeliveryTag uint64
}


var FieldMaps *map[string]string

func init() {
    FieldMaps = &map[string] string {
        "Method": "method",
        "Url": "url",
        "Headers": "headers",
        "Data": "data",
        "Params": "params",
        "Cookies": "cookies",
        "Timeout": "timeout",
        "ProxyName": "proxy_name",
        "CrawlerName": "crawler_name",
    }
}

func (r HttpRequest) ToJson() (string, error) {
    content, err := json.Marshal(r)
    if err != nil {
        return string(content), err
    }
    return FieldLower(string(content), false, *FieldMaps)
}


func FieldLower(v string, reverse bool, fieldMap map[string]string) (string, error) {
    var lower map[string] interface{}
    var upper map[string] interface{}
    var ret []byte
    var err error
    if reverse {
        err = json.Unmarshal([]byte(v), &lower)
        if err != nil {
            return string(ret), err
        }
        upper = make(map[string]interface{})
        for upperk, lowerk := range fieldMap {
            upper[upperk] = lower[lowerk]
        }
        ret, err = json.Marshal(upper)
    } else {
        err := json.Unmarshal([]byte(v), &upper)
        if err != nil {
            return string(ret), nil
        }
        lower = make(map[string]interface{})
        for upperk, lowerk := range fieldMap {
            lower[lowerk] = upper[upperk]
        }
        ret, err = json.Marshal(lower)
    }
    return string(ret), err
}

func FromJson(v string) (HttpRequest, error){
    var r HttpRequest
    newVal, err := FieldLower(v, true, *FieldMaps)
    if err != nil {
        return r, err
    }
    if err := json.Unmarshal([]byte(newVal), &r); err != nil {
        return r, err
    }
    return r, nil
}


func Push(ch *amqp.Channel, r HttpRequest, queueName string) error {
    v, err := r.ToJson()
    if err != nil {
        return err
    }
    msg := amqp.Publishing {
        DeliveryMode: amqp.Persistent,
        ContentType:  "text/plain",
        Body:         []byte(v),
    }
    err = ch.Publish("", queueName, false, false, msg)
    return err
}

func Done(ch *amqp.Channel, r *HttpRequest, success bool) error {
    var err error
    if r.DeliveryTag == 0 {
        return errors.New("HttpRequest DeliveryTag empty")
    }
    if success {
        err = ch.Ack(r.DeliveryTag,false)
    } else {
        err = ch.Nack(r.DeliveryTag,false, true)
    }
    return err
}

func Consume(ch *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
    deliveries, err := ch.Consume(
        queueName,
        "",
        false,
        false,
        false,
        false,
        nil,
    )
    return deliveries, err
}

func Get(d amqp.Delivery) (*HttpRequest, error) {
    var r HttpRequest
    if string(d.Body) == "" {
        return nil, nil
    }
    r, err := FromJson(string(d.Body))
    if err != nil {
        return &r, err
    }
    r.DeliveryTag = d.DeliveryTag
    return &r, err
}

