package model

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/titan/tingchao/utils"
)

// PackageModel
//
//	与客户端进行传递的包对象,用于封装二进制传输的消息列表
type PackageModel struct {
	MsgList []*MessageModel
}

func init() {
	gob.Register(&PackageModel{})
}

func NewPackageModel() *PackageModel {
	pm := &PackageModel{
		MsgList: make([]*MessageModel, 0),
	}
	return pm
}

// AddMsg
//
//	@description  : 为包对象添加一个消息对象，可以多次调用
//	@param         {MessageModel} msg	消息对象
func (p *PackageModel) AddMsg(msgID uint8, data interface{}) {
	msg := NewMessageModel1(msgID, data)
	p.MsgList = append(p.MsgList, msg)
}

// AddMsg1
//
//	添加消息对象,可以多次调用
func (p *PackageModel) AddMsg1(msg *MessageModel) {
	p.MsgList = append(p.MsgList, msg)
}

// EncodePackage  编码
//
//	@description  :将封包对象经过gob序列化并AES加密为二进制数组
//	@return        {*[]byte}		完成后的二进制数组地址
func (p *PackageModel) EncodePackage() (*[]byte, error) {
	// 序列化,方便网络传输
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(p) // gob序列化
	if err != nil {
		log.Println("[PackageModel.EncodePackage2GobAes]gob序列化错误", err)
		return nil, err
	}
	dataGob := result.Bytes()
	dataAes := utils.AESEncrypt(dataGob, []byte(utils.GlobalConfig.MessageAesKey)) // AES加密
	return &dataAes, nil
}

// DecodePackage  解码
//
//	@description  :将二进制数组经过gob反序列化并AES解密为消息数组
//	@param         {*[]byte} data	二进制数组
func (p *PackageModel) DecodePackage(data *[]byte) error {
	dataAes := utils.AESDecrypt(*data, []byte(utils.GlobalConfig.MessageAesKey)) // AES解密
	decoder := gob.NewDecoder(bytes.NewReader(dataAes))                          // gob解码
	err := decoder.Decode(p)
	if err != nil {
		log.Print("[DecodePackageList2GobAes]反序列化解码错误", err)
		return err
	}
	return nil
}
