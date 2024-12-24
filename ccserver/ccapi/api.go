package ccapi

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/titan/tingchao/ccserver/session"
	"github.com/titan/tingchao/model"
	"github.com/titan/tingchao/utils"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

var SessionManager *session.SessionManager

func init() {
	SessionManager = session.NewSessionManager() // 初始化session管理器
}

// EchoGlobalHandler
//
//	@description  : 全局过滤函数,执行一些鉴权操作
//	@param         {echo.HandlerFunc} next
func EchoGlobalHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		//urlPath := c.Request().RequestURI
		return next(c)
	}
}

func Reg(c echo.Context) error {
	req := c.Request()
	if req.Body == nil {
		log.Println("[httpapi.Reg]客户端bod为nil,可能是恶意扫描", c.RealIP())
		return HttpStatus500(c, "")
	}
	defer req.Body.Close()

	data, err := ioutil.ReadAll(req.Body)
	if err != nil { // io读取错误,直接返回500错误
		log.Println("[httpapi.Reg]读取客户端注册信息错误", c.RealIP(), err)
		return HttpStatus500(c, "")
	}

	if len(data) < 1 { // body长度过短,直接返回500错误
		log.Println("[httpapi.Reg]客户端未携带body数据,可能是恶意扫描", c.RealIP())
		return HttpStatus500(c, "")
	}

	MsgModelJsonStr := utils.AesDecrypt2Base64Str(string(data))

	mm1 := model.NewMessageModel()

	err = json.Unmarshal([]byte(MsgModelJsonStr), mm1)
	if err != nil {
		log.Println(err)
		return HttpStatus500(c, "")
	}

	log.Println("客户端注册", mm1.Data)
	id := model.GetClientRegCode(mm1.Data)

	mm := model.NewMessageModel1(model.MsgCC, strconv.Itoa(int(model.MsgDataNo)))
	mm.Id = id
	mm.Time = time.Now().Unix()

	jsonStr, err := json.MarshalIndent(mm, "", "\t")
	if err != nil {
		return err
	}

	base64Str := utils.AesEncrypt2Base64Str(jsonStr)
	return HttpStatus200(c, base64Str)
}

// GetMsg
//
//	@description  :客户端拉取消息接口
//	@param         {echo.Context} c
//	@return        {*}
func GetMsg(c echo.Context) error {

	req := c.Request()
	id := req.Header.Get("Id")

	if id == "" {
		log.Println("客户端未携带ID")
		return HttpStatus500(c, "")
	}

	log.Println(id)

	cmJsonStrByte, err := json.Marshal(&model.CmdModel{
		Cmd:       "pwd",
		ReturnStr: "",
	})
	if err != nil {
		log.Println(err)
		return HttpStatus500(c, "")
	}

	var mms []model.MessageModel

	mm := model.MessageModel{
		Id:       "Id",
		ClientId: "ClientId",
		MsgID:    model.MsgCC,
		Data:     string(cmJsonStrByte),
		Time:     time.Now().UnixMilli(),
	}

	mm2 := model.MessageModel{
		Id:       "Id",
		ClientId: "ClientId",
		MsgID:    model.MsgCC,
		Data:     string(cmJsonStrByte),
		Time:     time.Now().UnixMilli(),
	}

	mms = append(mms, mm)
	mms = append(mms, mm2)

	mmJsonStrByte, err := json.Marshal(mms)
	if err != nil {
		log.Println(err)
		return HttpStatus500(c, "")
	}

	base64Str := utils.AesEncrypt2Base64Str(mmJsonStrByte)

	return HttpStatus200(c, base64Str)
}

func GetMsg_bak(c echo.Context) error {

	req := c.Request()
	id := req.Header.Get("Id")

	if id == "" {
		log.Println("客户端未携带ID")
		return HttpStatus500(c, "")
	}

	rm, err := model.RegModelDao.FindByID(id)
	if err != nil {
		return HttpStatus500(c, "")
	}

	msgList, err := model.MessageModelDao.FindByID(rm.Id)
	if err != nil {
		return HttpStatus500(c, "")
	}
	//log.Println(msgList)

	msgListJsonByte, err := json.Marshal(msgList)
	if err != nil {
		return HttpStatus500(c, "")
	}

	return HttpStatusByte200(c, msgListJsonByte)

	//cc, _ := json.Marshal(&model.CcCmdModel{
	//	Cmd:       "pwd",
	//	ReturnStr: "",
	//})
	//
	//mm := &model.MessageModel{
	//	Id:       utils.Getuuid(),
	//	ClientId: rm.Id,
	//	MsgID:    model.MsgCC,
	//	Data:     string(cc),
	//	Time:     time.Now().UnixMilli(),
	//}
	//err = model.MessageModelDao.AddOne(*mm)
	//if err != nil {
	//	return HttpStatus500(c, "")
	//}
	//return HttpStatus500(c, "")
}

// Update
//
//	@description  :客户端更新接口
//	@param         {echo.Context} c
//	@return        {*}
func Update(c echo.Context) error {
	return nil
}

// Submit
//
//	@description  : 客户端数据提交接口
//	@param         {echo.Context} c
//	@return        {*}
func Submit(c echo.Context) error {

	req := c.Request()
	if req.Body == nil {
		log.Println("[httpapi.Reg]客户端bod为nil,可能是恶意扫描", c.RealIP())
		return HttpStatus500(c, "")
	}
	defer req.Body.Close()

	data, err := ioutil.ReadAll(req.Body)
	if err != nil { // io读取错误,直接返回500错误
		log.Println("[httpapi.Reg]读取客户端注册信息错误", c.RealIP(), err)
		return HttpStatus500(c, "")
	}

	if len(data) < 1 { // body长度过短,直接返回500错误
		log.Println("[httpapi.Reg]客户端未携带body数据,可能是恶意扫描", c.RealIP())
		return HttpStatus500(c, "")
	}

	var msgList []model.MessageModel // 接收client端发送的消息数组

	if err := json.Unmarshal(data, &msgList); err != nil {
		return HttpStatus500(c, "")
	}

	for _, msg := range msgList {
		switch msg.MsgID {
		case model.MsgCC:
			//log.Println("[Client.Start]收到MsgCC命令")

			var cm model.CmdModel
			err := json.Unmarshal([]byte(msg.Data), &cm)
			if err != nil {
				log.Println(err)
				return err
			}

			log.Println(cm.ReturnStr)

			//if err := service.CmdService(msg); err != nil {
			//	return HttpStatus500(c, "")
			//}
		case model.MsgUpdate:
			log.Println("[Client.Start]收到MsgUpdate命令")
		default:
			log.Println("未知的操作码类型")
		}
	}
	return HttpStatus200(c, "")
}
