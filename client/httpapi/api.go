package httpapi

import (
	"bytes"
	"crypto/tls"
	"github.com/titan/tingchao/utils"
	"log"
	"net/http"
	"time"
)

// Get getUrl:get访问的链接地址	header:请求头map,如果为nil则为默认值	isProxy:是否使用代理,默认使用 socks5://127.0.0.1:1080
//
//	url:请求url
//	header:请求头,如果为nil则使用默认请求头
//	timeOut:超时时间
func Get(url string, header map[string][]string, timeOut time.Duration, bodyByteData []byte) (*http.Response, error) {
	if header == nil {
		header = map[string][]string{
			"User-Agent": {"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)"},
			//"Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			//"Accept-Encoding": {"gzip, deflate, br"},
			//"Pragma":          {"no-cache"},
			"Content-Type":  {"application/octet-stream"},
			"Cache-Control": {"no-cache"},
			//"Connection":    {"close"},	// 客户端不看这个值
		}
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 跳过证书验证
		// 主动关闭与服务端的连接,可以让服务器少一些长连接,增强服务器性能
		// 对客户端性能可能会有影响,可能导致客户端端口被占满,主动关闭的一方tcp会出现time_wait状态
		// 一般持续2个60秒,如果一直关闭的话,会把客户端端口占完
		//DisableKeepAlives: true,
	}

	client := &http.Client{Timeout: timeOut * time.Second, Transport: transport} // 设置http超时时间和Transport对象

	bodyData := bytes.NewReader(bodyByteData)
	req, err := http.NewRequest("GET", url, bodyData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header = header
	rep, err := client.Do(req) // 访问,获得response对象

	defer client.CloseIdleConnections()

	return rep, err
}

// Post getUrl:get访问的链接地址	header:请求头map,如果为nil则为默认值	isProxy:是否使用代理,默认使用 socks5://127.0.0.1:1080
func Post(url string, header map[string][]string, timeOut time.Duration, bodyByteData []byte) (*http.Response, error) {
	if header == nil {
		header = map[string][]string{
			"User-Agent": {"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)"},
			//"Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			//"Accept-Encoding": {"gzip, deflate, br"},
			//"Pragma":          {"no-cache"},
			"Content-Type":  {"application/octet-stream"},
			"Cache-Control": {"no-cache"},
			//"Connection":    {"close"},	// 客户端不看这个值
		}
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 跳过证书验证
		// 主动关闭与服务端的连接,可以让服务器少一些长连接,增强服务器性能
		// 对客户端性能可能会有影响,可能导致客户端端口被占满,主动关闭的一方tcp会出现time_wait状态
		// 一般持续2个60秒,如果一直关闭的话,会把客户端端口占完
		//DisableKeepAlives: true,
	}

	client := &http.Client{Timeout: timeOut * time.Second, Transport: transport} // 设置http超时时间和Transport对象

	bodyData := bytes.NewReader(bodyByteData)
	req, err := http.NewRequest("POST", url, bodyData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header = header
	rep, err := client.Do(req) // 访问,获得response对象

	return rep, err
}

func ParseUrl(url string) string {
	domainList := utils.GetDomain()
	for _, domain := range domainList {
		log.Println(domain)
	}
	return ""
}
