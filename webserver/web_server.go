package webserver

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/titan/tingchao/utils"
	"github.com/titan/tingchao/webserver/webapi"
)

type WebServer struct {
}

func NewWebServer() *WebServer {
	ws := &WebServer{}
	return ws
}

func (w *WebServer) Run() {
	go func() {
		w.Start()
	}()
}

func (w *WebServer) Start() {
	e := echo.New()

	//e.Use(middleware.Logger()) // 设置日志
	e.HideBanner = true        // 隐藏Banner
	e.Static("/", "webstatic") //配置静态文件路径

	e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) { // 增加系统认证
		if username == utils.GlobalConfig.WebUserName && password == utils.GlobalConfig.WebUserPasswd {
			return true, nil
		}
		return false, nil
	}))

	//e.GET("/", webapi.Index)                             // 主页
	e.GET("/allconn", webapi.GetAllConn)                 // 获取所有链接
	e.GET("/killconn", webapi.KillConnByID)              // 根据id杀死链接
	e.GET("/execcmd", webapi.ExecCmd)                    // 执行命令
	e.GET("/getmsgistorybyid", webapi.GetMsgHistoryById) // 获取历史执行命令
	e.GET("/client_count", webapi.ClientCount)           // 客户端数据统计

	e.Logger.Fatal(e.Start(":" + utils.GlobalConfig.WebServerPort))
}
