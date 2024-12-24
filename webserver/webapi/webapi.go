package webapi

import (
	"github.com/labstack/echo/v4"
	"github.com/titan/tingchao/ccserver/api"
	"github.com/titan/tingchao/ccserver/session"
	"github.com/titan/tingchao/model"
	"github.com/titan/tingchao/webserver/httpmodel"
	"log"
	"net/http"
	"net/url"
	"time"
)

func Index(c echo.Context) error {
	return c.String(http.StatusOK, "web server index")
}

// GetAllConn 获取所有连接信息
func GetAllConn(c echo.Context) error {

	connList := httpmodel.ConnList{
		Code:       200,
		Msg:        "ok",
		ClientList: make([]httpmodel.Client, 0),
	}

	// 此处必须对sessionMap加锁，否则会出现并发错误
	storage := ccapi.SessionManager.Storage.(*session.StorageFromMemory)
	storage.Lock.RLock()
	defer storage.Lock.RUnlock()

	pSessionMap := storage.SessionMap

	for _, session := range pSessionMap {

		pClientModel := session.Get("client_model").(*model.ClientModel)

		client := httpmodel.Client{
			SessionID:        url.QueryEscape(session.GetId()),
			IPAddress:        pClientModel.IPAddress,
			Version:          pClientModel.ClientRegModel.ClientVersion,
			LastAccessedTime: session.GetLastAccessedTime().Unix(), // 固定格式化字符串...
			Uid:              pClientModel.ClientRegModel.Uid,
			Gid:              pClientModel.ClientRegModel.Gid,
			Username:         pClientModel.ClientRegModel.Username,
			Name:             pClientModel.ClientRegModel.Name,
			HomeDir:          pClientModel.ClientRegModel.HomeDir,
			SystemType:       pClientModel.ClientRegModel.SystemType,
			SystemArch:       pClientModel.ClientRegModel.SystemArch,
			Hostname:         pClientModel.ClientRegModel.Hostname,
		}
		connList.ClientList = append(connList.ClientList, client)
	}
	return c.JSON(http.StatusOK, connList)
}

func KillConnByID(c echo.Context) error {
	log.Println("===> 收到前端命令关闭此conn连接, ID=", c.QueryParam("id"))

	sessionID, err := url.QueryUnescape(c.QueryParam("id"))
	if err != nil {
		log.Println("[webapi.KillConnByID]url转码错误")
		return c.String(http.StatusInternalServerError, "500")
	}

	pSession := ccapi.SessionManager.GetSessionById(sessionID)
	if pSession == nil {
		log.Println("[webapi.KillConnByID]未找到Session对象")
		return c.String(http.StatusInternalServerError, "500")
	}

	pClientModel := pSession.Get("client_model").(*model.ClientModel)
	if pClientModel == nil {
		log.Println("[webapi.KillConnByID]未找到ClientModel对象")
		return c.String(http.StatusInternalServerError, "500")
	}

	pMsg := model.NewMessageModel1(model.MsgCloseClient, nil)
	pClientModel.AddServerMsg(pMsg) // 添加关闭客户端消息

	return c.String(http.StatusOK, "")
}

func ExecCmd(c echo.Context) error {
	log.Println("==>[web ExecCmd] , id=", c.QueryParam("id"), "cmd=", c.QueryParam("cmd"))

	sessionID := c.QueryParam("id")
	cmd := c.QueryParam("cmd")

	pSession := ccapi.SessionManager.GetSessionById(sessionID)
	if pSession == nil {
		log.Println("[webapi.KillConnByID]未找到Session对象")
		return c.String(http.StatusInternalServerError, "500")
	}

	pClientModel := pSession.Get("client_model").(*model.ClientModel)
	if pClientModel == nil {
		log.Println("[webapi.KillConnByID]未找到ClientModel对象")
		return c.String(http.StatusInternalServerError, "500")
	}

	pCm := model.NewCcCmdModel()
	pCm.Cmd = cmd

	pMsg := model.NewMessageModel1(model.MsgCC, pCm)
	pClientModel.AddServerMsg(pMsg) // 添加执行消息命令

	return c.String(http.StatusOK, "")
}

func GetMsgHistoryById(c echo.Context) error {

	//log.Println("==>[web GetMsgHistoryById] , id=", c.QueryParam("id"))

	sessionID := c.QueryParam("id")

	pSession := ccapi.SessionManager.GetSessionById(sessionID)
	if pSession == nil {
		log.Println("[webapi.KillConnByID]未找到Session对象")
		return c.String(http.StatusInternalServerError, "500")
	}

	pClientModel := pSession.Get("client_model").(*model.ClientModel)
	if pClientModel == nil {
		log.Println("[webapi.KillConnByID]未找到ClientModel对象")
		return c.String(http.StatusInternalServerError, "500")
	}

	cmdList := httpmodel.CmdList{
		Code: 200,
		Cc:   make([]model.CcCmdModel, 0),
		Msg:  "ok",
	}

	for _, messageModel := range pClientModel.ClientMsgList {
		if messageModel.MsgID == model.MsgCC {
			pCm := messageModel.Data.(*model.CcCmdModel)
			cmdList.Cc = append(cmdList.Cc, *pCm)
		}
	}

	return c.JSON(http.StatusOK, cmdList)
}

func ClientCount(c echo.Context) error {

	storage := ccapi.SessionManager.Storage.(*session.StorageFromMemory)
	storage.Lock.RLock()
	defer storage.Lock.RUnlock()

	pSessionMap := storage.SessionMap

	clientCount := len(pSessionMap)
	clientCount10Min := 0
	clientCount1Hour := 0
	clientCount24Hour := 0
	linuxClientCount := 0
	winClientCount := 0
	otherClientCount := 0

	timeUnix := time.Now().Unix() // 老版本GO使用此cpi会报错,老版本GO time类不存在获取毫秒的方法(UnixMilli)
	for _, session := range pSessionMap {

		pClientModel := session.Get("client_model").(*model.ClientModel)

		if pClientModel.ClientRegModel.SystemType == "linux" {
			linuxClientCount += 1
		} else if pClientModel.ClientRegModel.SystemType == "windows" {
			winClientCount += 1
		} else {
			otherClientCount += 1
		}

		time1 := timeUnix - session.GetLastAccessedTime().Unix()

		if time1 <= 600 { // 10分钟内活跃的客户端
			clientCount10Min += 1
		}
		if time1 <= 3600 { // 1小时内活跃的客户端
			clientCount1Hour += 1
		}
		if time1 <= 86400 { // 24小时内活跃的客户端
			clientCount24Hour += 1
		}
	}

	clientCountModel := &httpmodel.ClientCountModel{
		Code:              200,
		Msg:               "ok",
		ClientCount:       clientCount,
		ClientCount10Min:  clientCount10Min,
		ClientCount1Hour:  clientCount1Hour,
		ClientCount24Hour: clientCount24Hour,
		LinuxClientCount:  linuxClientCount,
		WinClientCount:    winClientCount,
		OtherClientCount:  otherClientCount,
	}
	return c.JSON(http.StatusOK, clientCountModel)
}
