package ccserver

import (
	"fmt"
	"github.com/titan/tingchao/ccserver/ccapi"
	"github.com/titan/tingchao/ccserver/common"
)

type CoreServer struct{}

func NewCoreServer() *CoreServer {
	coreServer := &CoreServer{}
	return coreServer
}

func (c *CoreServer) Run() {

	fmt.Printf(common.Banner, common.Version)

	e := ccapi.InitEcho()

	// 一些初始化操作防在这

	e.Logger.Fatal(e.Start(":" + common.Configs.CcServerPort))
}
