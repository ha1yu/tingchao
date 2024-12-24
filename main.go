package main

import (
	"github.com/titan/tingchao/ccserver"
	"github.com/titan/tingchao/webserver"
)

func main() {

	// 开启cc服务
	cs := ccserver.NewCcServer()
	cs.Run()

	// 开启web管理服务
	ws := webserver.NewWebServer()
	ws.Run()

	select {}
}
