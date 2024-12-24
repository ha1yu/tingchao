package httpmodel

import "github.com/titan/tingchao/model"

type CmdList struct {
	Code int16              // json状态吗
	Cc   []model.CcCmdModel // 命令对象
	Msg  string             // json消息
}
