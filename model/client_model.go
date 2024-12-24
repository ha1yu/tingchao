package model

import (
	"sync"
)

type ClientModel struct {
	Lock           sync.Mutex      // 锁
	IPAddress      string          // 客户端ip地址
	ClientRegModel *RegModel       // 客户端注册消息
	ServerMsgList  []*MessageModel // 服务器等待发送给客户端的消息列表
	ClientMsgList  []*MessageModel // 收到的客户端发送给服务器的消息列表
}

func NewClientModel() *ClientModel {
	cm := &ClientModel{
		Lock:           sync.Mutex{},
		ClientRegModel: nil,
		ServerMsgList:  make([]*MessageModel, 0),
		ClientMsgList:  make([]*MessageModel, 0),
	}
	return cm
}

// AddServerMsg 添加待发送消息
func (c *ClientModel) AddServerMsg(msg *MessageModel) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if len(c.ServerMsgList) < 5 { // 列表不满,直接放进去
		c.ServerMsgList = append(c.ServerMsgList, msg)
	} else { // 列表满了,此时删除第一个,然后再写入新的
		c.ServerMsgList = c.ServerMsgList[1:]
		c.ServerMsgList = append(c.ServerMsgList, msg)
	}
}

// AddClientMsg 添加客户端消息
func (c *ClientModel) AddClientMsg(msg *MessageModel) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if len(c.ClientMsgList) < 5 { // 列表不满,直接放进去
		c.ClientMsgList = append(c.ClientMsgList, msg)
	} else { // 列表满了,此时删除第一个,然后再写入新的
		c.ClientMsgList = c.ClientMsgList[1:]
		c.ClientMsgList = append(c.ClientMsgList, msg)
	}
}

// GetAndRemoveServerMsg 获取所有的消息,并清空消息列表
func (c *ClientModel) GetAndRemoveServerMsg() *[]*MessageModel {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	msgList := make([]*MessageModel, len(c.ServerMsgList))
	copy(msgList, c.ServerMsgList)             // 复制一份数据返回
	c.ServerMsgList = make([]*MessageModel, 0) // 清空消息列表
	return &msgList
}
