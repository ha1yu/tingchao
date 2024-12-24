package model

import "encoding/gob"

// cc模型

type CcCmdModel struct {
	Cmd       string // 待执行的命令
	ReturnStr string //命令执行的结果，返回给服务端
}

func init() {
	gob.Register(&CcCmdModel{})
}

func NewCcCmdModel() *CcCmdModel {
	c := &CcCmdModel{
		Cmd:       "",
		ReturnStr: "",
	}
	return c
}
