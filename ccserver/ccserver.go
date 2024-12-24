/*
 * @Author       : hai
 * @Date         : 2021-08-30 23:00:50
 * @LastEditTime : 2021-09-20 19:28:35
 * @LastEditors  : hai
 * @Description  : hai
 * @FilePath     : /duck-cc-server-http/ccserver/ccserver.go
 * Copyright 2021 <hai>, All Rights Reserved
 */

package ccserver

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/titan/tingchao/ccserver/api"
	"github.com/titan/tingchao/utils"
	"log"
	"net/http"
)

type CcServer struct {
}

func NewCcServer() *CcServer {
	ccServer := &CcServer{}
	return ccServer
}

func (c *CcServer) Run() {
	go func() {
		c.Start()
	}()
}

func (c *CcServer) Start() {
	e := echo.New()

	//e.Use(middleware.Logger())         // 设置日志
	//e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}))// 开启gzip压缩,开启可能会影响性能
	//e.Use(ccapi.EchoGlobalHandler)    // 添加全局过滤函数,执行一些鉴权操作

	e.Use(middleware.BodyLimit("2M")) // 限制客户端发送的body数据长度为2兆

	e.HTTPErrorHandler = MyHTTPErrorHandler // 自定义错误,echo自带的404错误处理默认返回json,影响服务器性能
	e.HideBanner = true                     // 隐藏Banner

	e.Static("/", "ccstatic") //配置静态文件路径

	e.POST("/reg", ccapi.Reg)       // 客户端注册路由
	e.GET("/get", ccapi.GetMsg)     // 客户端拉取消息路由
	e.GET("/min", ccapi.Mining)     // 客户端获取矿工信息路由
	e.GET("/update", ccapi.Update)  // 客户端更新路由
	e.POST("/sub", ccapi.Submit)    // 客户端提交数据路由
	e.GET("/b6u2t91jv", ccapi.Reg1) // 测试路由

	e.Logger.Fatal(e.Start(":" + utils.GlobalConfig.CcServerPort))
}

// MyHTTPErrorHandler 自定义错误处理方法
//
//	错误信息直接返回一个只包含http状态码的响应
//	echo默认的错误处理会返回一个标准的RESTful API 一个json格式的错误响应
//	如果有大量的扫描服务器会影响服务器性能,将其替换为返回一个只包含http状态码的响应
//	HTTP/1.1 404 Not Found
//	Date: Thu, 23 Sep 2021 07:31:37 GMT
//	Content-Length: 0
func MyHTTPErrorHandler(err error, c echo.Context) {
	he, ok := err.(*echo.HTTPError)
	if ok {
		if he.Internal != nil {
			if herr, ok := he.Internal.(*echo.HTTPError); ok {
				he = herr
			}
		}
	} else {
		he = &echo.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		}
	}

	code := he.Code
	message := he.Message

	// Send response
	if !c.Response().Committed {
		//err = c.String(code, "")
		log.Println(c.RealIP(), c.Request().RequestURI, code, message)
		err = c.NoContent(he.Code)
	}
}
