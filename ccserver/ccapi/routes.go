package ccapi

import "github.com/labstack/echo/v4"

func AddRoutes(e *echo.Echo) {

	e.Static("/", "ccstatic") //配置静态文件路径

	//e.POST("/reg", ccapi.Reg)       // 客户端注册路由
	//e.GET("/get", ccapi.GetMsg)     // 客户端拉取消息路由
	//e.GET("/update", ccapi.Update)  // 客户端更新路由
	//e.POST("/sub", ccapi.Submit)    // 客户端提交数据路由
	//e.GET("/b6u2t91jv", ccapi.Reg1) // 测试路由

	e.POST("/reg", Reg)    // 客户端注册路由
	e.GET("/get", GetMsg)  // 客户端拉取消息路由
	e.POST("/sub", Submit) // 客户端提交数据路由
}
