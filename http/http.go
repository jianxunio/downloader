package http

import "downloader/request"
import "downloader/response"
import "downloader/proxy"
import "downloader/ssdb"
import "net/http"
import "net/url"
import "time"
import "fmt"
import "strings"
import "io/ioutil"
import "errors"

func getProxy(proxyClient *ssdb.ProxyClient, proxyName string, reqUrl string) (string, error){
    var err error
    p := ""
    url := ""
    if proxyName != "" {
        p, err = proxy.GetOne(proxyClient,proxyName)
        if err != nil {
            return url, errors.New(fmt.Sprintf("proxy.GetOne fail %s", err.Error()))
        }
        if strings.HasPrefix(reqUrl, "https") {
            url = fmt.Sprintf("https://%s", p)
        } else if strings.HasPrefix(reqUrl, "http") {
            url = fmt.Sprintf("http://%s", p)
        } else {
            return url,errors.New(fmt.Sprintf("reqUrl %s not valid", reqUrl))
        }
        return url, err
    }
    return url, err
}

func customProxy(req *http.Request) (*url.URL, error) {
    httpProxy := req.Header.Get("HttpProxy")
    if httpProxy == "" {
        return nil, nil
    } else {
        return url.Parse(httpProxy)
    }
}

func GetClient() *http.Client {
    var client http.Client
    client = http.Client{
        Timeout: 15 * time.Second,
    }
    transport :=  http.Transport{
        Proxy: customProxy,
        DisableCompression: false,
        MaxIdleConnsPerHost: 4000,
        /*
        Dial: (&net.Dialer{
                    Timeout:   15 * time.Second,
                    KeepAlive: 15 * time.Second,
                }).Dial,
        */
        TLSHandshakeTimeout: 5 * time.Second,
    }
    client.Transport = &transport
    return &client
}

func structResponse(req *request.HttpRequest, content *http.Response, proxy string) (response.HttpResponse, error) {
    var resp response.HttpResponse
    data, err := ioutil.ReadAll(content.Body)
    if err != nil {
        return resp, err
    }
    header := make(map[string]interface{})
    for k, v := range content.Header {
        if len(v) > 0 {
            header[k] = v[0]
        }
    }
    var encoding string
    encoding = content.Header.Get("Content-Encoding")
    resp = response.HttpResponse {
        Request: *req,
        ErrorCode: 0,
        ErrorMsg: "",
        StatusCode: content.StatusCode,
        Reason: content.Status,
        Html: string(data),
        Headers: header,
        Encoding: encoding,
        Url: req.Url,
        CrawlerName: req.CrawlerName,
        ProxyName: req.ProxyName,
        HttpProxy: proxy,
    }
    return resp, err
}

func Send(proxyClient *ssdb.ProxyClient, r *request.HttpRequest, httpClient *http.Client) (response.HttpResponse, error) {
    var err error
    var httpProxy string
    resp := response.HttpResponse{}

    httpProxy, err = getProxy(proxyClient, r.ProxyName, r.Url)
    if err != nil {
        return resp, err
    }
    resp.HttpProxy = httpProxy

    req, err := http.NewRequest(r.Method, r.Url, strings.NewReader(r.Data))
    if err != nil {
        return resp, err
    }

    u, err := url.Parse(r.Url)
    if err != nil {
        return resp, err
    }

    defaultHeaders := map[string] string {
        // "User-Agent": `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_1) 
        // AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36`,
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36",
        "Upgrade-Insecure-Requests": "1",
        "Connection": "keep-alive",
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
        "Accept-Language": "zh-CN,zh;q=0.8,en;q=0.6",
        "Cache-Control": "max-age=0",
        "Host": u.Host,
        "HttpProxy": httpProxy,
    }

    for k, v := range defaultHeaders {
        req.Header.Set(k, v)
    }

    for k, v := range r.Headers {
        req.Header.Set(k, fmt.Sprintf("%v", v))
    }

    /*for k, v := range r.Cookies {
        req.Header.Set("Cookie", fmt.Sprintf("%s=%v", k, v))
    }*/


    values := req.URL.Query()
    for k, v := range r.Params {
        values.Add(k, fmt.Sprintf("%v", v))
    }
    req.URL.RawQuery = values.Encode()

    content, err := httpClient.Do(req)
    if err != nil {
        return resp, err
    }
    defer content.Body.Close()
    resp, err = structResponse(r, content, httpProxy)
    if err != nil {
        return resp, err
    }
    respCookies := make(map[string] interface{})
    if httpClient.Jar != nil {
        cookies := httpClient.Jar.Cookies(req.URL)
        for _, cookie := range cookies {
            respCookies[cookie.Name] = cookie.Value
        }
        resp.Cookies = respCookies
    }
    return resp, err
}
