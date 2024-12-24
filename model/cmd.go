package model

import (
	"encoding/json"
	"log"
)

// cc模型

type CmdModel struct {
	Cmd       string // 待执行的命令
	ReturnStr string //命令执行的结果,返回给服务端
}

func NewCmdModel() *CmdModel {
	c := &CmdModel{
		Cmd:       "",
		ReturnStr: "",
	}
	return c
}

func (c CmdModel) String() string {
	str, err := json.Marshal(c)
	if err != nil {
		log.Println(err)
	}
	return string(str)
}
