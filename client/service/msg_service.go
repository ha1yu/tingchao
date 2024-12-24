package service

import (
	"errors"
	"fmt"
	"github.com/titan/tingchao/client/httpapi"
	"io/ioutil"
	"log"

	"github.com/titan/tingchao/model"
	"github.com/titan/tingchao/utils"
)

// GetMsgListFromServer 从服务端拉取待执行的任务
func GetMsgListFromServer() ([]*model.MessageModel, error) {
	//url := "http://127.0.0.1:8080/get"
	url := fmt.Sprintf("http://%s:%s/get", utils.GlobalConfig.ServerHost, utils.GlobalConfig.ServerPort)

	header := map[string][]string{
		"User-Agent":    {"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)"},
		"cookie":        {utils.GlobalConfig.Cookie.Name + "=" + utils.GlobalConfig.Cookie.Value},
		"Cache-Control": {"no-cache"},
		"Connection":    {"close"},
	}

	rep, err := httpapi.Get(url, header, utils.GlobalConfig.HttpTimeOut, nil)
	if err != nil {
		log.Println("[api.GetMsgListFromServer]请求错误", err)
		return nil, err
	}
	defer rep.Body.Close()

	if rep.StatusCode == 401 { // cookie过期,需要重新注册客户端
		log.Println("[api.GetMsgListFromServer]cookie过期,需要重新注册客户端", rep.StatusCode)
		return nil, errors.New("401")
	}
	if rep.StatusCode != 200 { // 没有成功请求到数据直接返回nil
		log.Println("[api.GetMsgListFromServer]请求错误,状态码", rep.StatusCode)
		return nil, errors.New("[api.GetMsgListFromServer]请求错误")
	}

	data, err := ioutil.ReadAll(rep.Body)
	if err != nil {
		log.Print("[api.GetMsgListFromServer]io读取错误", err)
		return nil, err
	}
	if len(data) < 1 { // 客户端返回200,而且没有body,代表没有消息
		log.Println("[api.GetMsgListFromServer]服务端没有消息")
		return nil, errors.New("[api.GetMsgListFromServer]服务端没有消息")
	}

	pm := model.NewPackageModel()
	err = pm.DecodePackage(&data)
	if err != nil {
		log.Print("[api.Reg]解码失败", err)
		return nil, err
	}

	//for _, messageModel := range pm.MsgList {
	//	log.Println(messageModel)
	//}

	pMsgList := pm.MsgList

	return pMsgList, nil
}
