package service

import (
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"time"

	"github.com/titan/tingchao/client/httpapi"

	"github.com/titan/tingchao/model"
	"github.com/titan/tingchao/utils"
)

type UpdateService struct {
}

func NewUpdateService() *UpdateService {
	return &UpdateService{}
}

// UpdateWatchService 客户端更新服务线程
// TODO MAC的自动更新有问题,待解决(从网上下载的文件,执行的时候会弹窗提示有风险)
func UpdateWatchService() {
	// 创建一个周期性的定时器,随机1-2天
	ticker := time.NewTicker(time.Duration(utils.RandomNumInt(60*60*24*1, 60*60*24*2)) * time.Second)
	go func() {
		for {
			<-ticker.C
			um := PullUpdateMessage() // 获取服务端的更新对象
			UpdateRun(um)
		}
	}()
}

// UpdateGo 执行一次客户端更新
func UpdateGo(msg *model.MessageModel) {
	um := PullUpdateMessage() // 获取服务端的更新对象
	UpdateRun(um)
}

// PullUpdateMessage 从服务端拉取Update信息
func PullUpdateMessage() *model.UpdateModel {
	//url := "http://127.0.0.1:8080/update"
	url := fmt.Sprintf("http://%s:%s/update", utils.GlobalConfig.ServerHost, utils.GlobalConfig.ServerPort)

	header := map[string][]string{
		"User-Agent":    {"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)"},
		"cookie":        {utils.GlobalConfig.Cookie.Name + "=" + utils.GlobalConfig.Cookie.Value},
		"Cache-Control": {"no-cache"},
		"Connection":    {"close"},
	}

	rep, err := httpapi.Get(url, header, utils.GlobalConfig.HttpTimeOut, nil)
	if err != nil { // 连接超时这一类的错误
		log.Println("[api.Update]请求错误", err)
		return nil
	}
	defer rep.Body.Close() // 使用完关闭连接body

	if rep.StatusCode != 200 { // 没有成功请求到数据直接返回nil
		log.Println("[api.Update]请求错误,状态码", rep.StatusCode)
		return nil
	}

	body, err := ioutil.ReadAll(rep.Body)
	if err != nil {
		log.Print("[api.Update]io读取错误", err)
		return nil
	}

	pm := model.NewPackageModel()
	err = pm.DecodePackage(&body)
	if err != nil {
		log.Print("[api.Update]解码失败", err)
		return nil
	}

	if _, ok := pm.MsgList[0].Data.(*model.UpdateModel); !ok {
		log.Print("[api.Update]类型转换错误")
		return nil
	}

	um := pm.MsgList[0].Data.(*model.UpdateModel)

	log.Println("[api.Update]服务端拉取Update信息成功")
	return um
}

// UpdateRun
//
//	@description  :更新客户端,完成下载执行的操作,下载失败或者启动失败会return此方法
//	@param         {*model.UpdateModel} um	更新model
func UpdateRun(um *model.UpdateModel) {
	if um == nil { // 连接出错会获取到nil
		log.Println("[UpdateRun]从服务器拉取数据失败")
		return
	}
	if utils.GlobalConfig.ClientVersion != um.ClientVersion { // 如果本地版本跟远程版本不一致,则更新客户端
		log.Println("[UpdateRun]客户端与服务端版本不一致,开始更新！")
		downPath := utils.GlobalConfig.ClientPath

		if len(um.UpdateUrl) > 0 { //下载更新文件
			switch runtime.GOOS {
			case "windows":
				if ok := utils.DownloadFile(um.UpdateUrl, downPath, utils.GlobalConfig.ClientNameWin); ok { //win下载成功
					log.Println("[UpdateRun]win 下载文件成功,继续执行软件; 下载路径: ", downPath)

					_, err := utils.ExecShell("start /b " + downPath + "/" + utils.GlobalConfig.ClientNameWin) //后台执行程序
					if err != nil {
						log.Println("[UpdateRun]start /b 执行出错,结束此次更新服务")
						return
					}

					// 更新成功之后再杀死客户端守护脚本
					// 未实现

					log.Println("[UpdateRun]新程序已启动,结束程序中...")

					utils.GlobalConfig.CloseClientType <- true // 更新程序运行成功之后,结束本进程
				} else {
					log.Println("[UpdateRun]windows文件下载失败")
				}
			default: // 除了win之外的下载tar.gz包
				if ok := utils.DownloadFile(um.UpdateUrl, downPath, utils.GlobalConfig.ClientNameUnix); ok { //下载成功
					log.Println("[UpdateRun]下载文件成功,继续执行软件; 下载路径: ", downPath)

					_, err := utils.ExecShell("chmod +x " + downPath + "/" + utils.GlobalConfig.ClientNameUnix)
					if err != nil {
						log.Println("[UpdateRun]chmod +x 命令执行错误,结束此次更新服务", err)
						return
					}

					_, err = utils.ExecShell("nohup " + downPath + "/" + utils.GlobalConfig.ClientNameUnix + " >/dev/null 2>&1 & ")
					if err != nil {
						log.Println("[UpdateRun]nohup 命令执行错误,结束此次更新服务", err)
						return
					}

					log.Println("[UpdateRun]新程序已启动,结束程序中...")

					utils.GlobalConfig.CloseClientType <- true // 更新程序运行成功之后,结束本进程
				} else {
					log.Println("[UpdateRun]default文件下载失败")
				}
			}
		} else {
			log.Println("[UpdateRun]um.UpdateUrl长度为0,跳过本次更新")
		}
	} else {
		log.Println("[UpdateRun]客户端与服务端版本一致,无需更新！")
	}
}
