package service

import (
	"fmt"
	"github.com/titan/tingchao/client/httpapi"
	"log"
	"time"

	"github.com/titan/tingchao/model"
	"github.com/titan/tingchao/utils"
)

// Reg 注册逻辑
//
//	必须把逻辑抽出去,不然[defer rep.Body.Close()]在死循环中无法执行,会导致内存泄漏
func Reg() {
	for {
		isok := regRun()
		if isok { // 注册成功直接退出注册函数
			return
		}
		time.Sleep(time.Duration(utils.GlobalConfig.TimeSleep) * time.Second)
	}
}

/**
 * @description  : 注册主逻辑,必须把逻辑抽出来,不然[defer rep.Body.Close()]在死循环中无法执行,会导致内存泄漏
 * @param         {*}
 * @return        {bool}	注册成功返回true 失败返回false
 */
func regRun() bool {
	//url := "http://127.0.0.1:8080/reg"
	url := fmt.Sprintf("http://%s:%s/reg", utils.GlobalConfig.ServerHost, utils.GlobalConfig.ServerPort)

	header := map[string][]string{
		"User-Agent":    {"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)"},
		"Content-Type":  {"application/octet-stream"},
		"Cache-Control": {"no-cache"},
		"Connection":    {"close"},
	}

	rm := model.NewRegModel()

	pm := model.NewPackageModel()
	pm.AddMsg(model.MsgReg, rm)

	data, err := pm.EncodePackage()
	if err != nil {
		log.Println("[api.Reg]gob序列化错误", err)
		time.Sleep(time.Duration(utils.GlobalConfig.TimeSleep) * time.Second)
		return false
	}

	rep, err := httpapi.Post(url, header, utils.GlobalConfig.HttpTimeOut, *data)
	if err != nil { // http出现请求出现错误,重新注册
		log.Println("[api.Reg]请求错误", err)
		time.Sleep(time.Duration(utils.GlobalConfig.TimeSleep) * time.Second)
		return false
	}
	defer rep.Body.Close()

	if rep.StatusCode != 200 { // 未成功请求到数据
		log.Println("[api.Reg]请求错误")
		return false
	}

	if len(rep.Cookies()) < 1 { // 未从服务端获取到cookie值,那么重新注册客户端
		log.Println("[api.Reg]未获取到cookie,正在重新获取")
		time.Sleep(time.Duration(utils.GlobalConfig.TimeSleep) * time.Second)
		return false
	}

	// 获取到cooike之后放入全局对象中
	for _, cookie := range rep.Cookies() {
		//log.Println("[api.Reg]客户端注册完成,获取到下发Cookie:\n", cookie)
		utils.GlobalConfig.Cookie = cookie
	}
	return true // 成功获取cookie之后继续执行后面的逻辑

}
