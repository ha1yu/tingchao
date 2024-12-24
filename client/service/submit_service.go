package service

import (
	"errors"
	"fmt"
	"github.com/titan/tingchao/client/httpapi"
	"github.com/titan/tingchao/utils"
	"log"
)

// Submit2Server 消息提交方法
//
//	传入待提交的二进制消息,发送给服务器的sub接口
//	传入的二进制数据需要先做加密和序列化的操作,此方法不做加密和序列化操作
func Submit2Server(byte []byte) error {
	//url := "http://127.0.0.1:8080/sub"
	url := fmt.Sprintf("http://%s:%s/sub", utils.GlobalConfig.ServerHost, utils.GlobalConfig.ServerPort)

	header := map[string][]string{
		"User-Agent":    {"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)"},
		"cookie":        {utils.GlobalConfig.Cookie.Name + "=" + utils.GlobalConfig.Cookie.Value},
		"Cache-Control": {"no-cache"},
		"Connection":    {"close"},
	}

	req, err := httpapi.Post(url, header, utils.GlobalConfig.HttpTimeOut, byte)
	if err != nil { // 连接超时这一类的错误
		log.Println("[service.Submit2Server]请求错误", err)
		return err
	}
	defer req.Body.Close() // 使用完关闭连接body

	if req.StatusCode != 200 { // 没有成功请求到数据直接返回nil
		log.Println("[service.Submit2Server]请求错误,状态码", req.StatusCode)
		return errors.New("no_200")
	}
	return nil
}
