package response

import "downloader/request"
import "encoding/json"
import "fmt"
import "downloader/ssdb"
import "github.com/streadway/amqp"
import "github.com/garyburd/redigo/redis"
import "errors"

type HttpResponse struct {
    Request request.HttpRequest
    RequestJson string
    ErrorCode  int
    ErrorMsg   string
    StatusCode int
    Reason string
    Html    string
    Cookies map[string] interface{}
    Url        string
    Headers map[string] interface{}
    Encoding string
    CrawlerName string
    ProxyName string
    HttpProxy string
}

var FieldMaps *map[string]string


func (r HttpResponse) ToJson() (string, error) {
    req_json,err := r.Request.ToJson()
    if err != nil {
        return "", err
    }
    r.RequestJson = req_json
    content, err := json.Marshal(r)
    if err != nil {
        return string(content), err
    }
    return request.FieldLower(string(content), false, *FieldMaps)
}


func init() {
    FieldMaps = &map[string] string {
        "RequestJson": "http_request",
        "ErrorCode": "error_code",
        "ErrorMsg": "error_msg",
        "StatusCode": "status_code",
        "Reason": "reason",
        "Html": "html",
        "Url": "url",
        "Headers": "headers",
        "Cookies": "cookies",
        "Encoding": "encoding",
        "CrawlerName": "crawler_name",
        "ProxyName": "proxy_name",
        "HttpProxy": "http_proxy",
    }
}


func PushKey(ch *amqp.Channel, r HttpResponse, queueName string) error {
    v := fmt.Sprintf("http_response:%s:%s", r.CrawlerName, r.Url)
    msg := amqp.Publishing {
        // DeliveryMode: amqp.Persistent,
        DeliveryMode: amqp.Transient,
        ContentType:  "text/plain",
        Body:         []byte(v),
    }
    err := ch.Publish("", queueName, false, false, msg)
    return err
}

func Push(ssdbClients []*ssdb.SSDBClient, r HttpResponse) error {
    k := fmt.Sprintf("http_response:%s:%s", r.CrawlerName, r.Url)
    client, err := ssdb.GetClient(ssdbClients, k)
    if err != nil {
        return errors.New(fmt.Sprintf("ssdb.GetClient error %s", err.Error()))
    }
    rc := client.ConnPool
    v, err := r.ToJson()
    if err != nil {
        return errors.New(fmt.Sprintf("ToJson error %s", err.Error()))
    }
    db  := rc.Get()
    defer db.Close()
    _, err = db.Do("set", redis.Args{}.AddFlat(k).AddFlat(v)...)
    if err != nil {
        return errors.New(fmt.Sprintf("set redis %s:%d error %s", client.Node.Host, client.Node.Port, err.Error()))
    }
    return nil
}
