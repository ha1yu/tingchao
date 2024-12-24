package model

import (
	"encoding/gob"

	"github.com/titan/tingchao/utils"
)

// 挖矿模型

type MiningModel struct {
	MiningUrl     string // 矿工下载地址
	PoolUrl       string // 矿池地址 配置文件的 url 值
	WalletAddress string // xmr钱包地址 配置文件的 user 值
	XmrigName     string // 矿工名字 配置文件的 pass 值
}

func init() {
	gob.Register(&MiningModel{})
}

func NewMiningModel1() *MiningModel {
	mm := &MiningModel{
		MiningUrl:     utils.GlobalConfig.MiningWindows386,
		PoolUrl:       utils.GlobalConfig.PoolUrl,
		WalletAddress: utils.GlobalConfig.WalletAddress,
		XmrigName:     utils.GlobalConfig.XmrigName,
	}
	return mm
}

/**
 * @description  :会根据clientModel中存储的注册信息中系统架构自动判断,应该下载的版本
 * @param         {*ClientModel} cm		客户端对象
 * @return        {*MiningModel}		矿工对象
 */
func NewMiningModel(cm *ClientModel) *MiningModel {

	mm := &MiningModel{
		MiningUrl:     utils.GlobalConfig.MiningWindows386,
		PoolUrl:       utils.GlobalConfig.PoolUrl,
		WalletAddress: utils.GlobalConfig.WalletAddress,
		XmrigName:     utils.GlobalConfig.XmrigName,
	}

	systemType := cm.ClientRegModel.SystemType
	systemArch := cm.ClientRegModel.SystemArch

	switch systemType { // 不同的系统不同架构下载不同的更新包
	case "windows":
		switch systemArch { // 分辨架构
		case "386":
			mm.MiningUrl = utils.GlobalConfig.MiningWindows386
		case "amd64":
			mm.MiningUrl = utils.GlobalConfig.MiningWindowsAmd64
		default:
			mm.MiningUrl = utils.GlobalConfig.MiningWindowsDefault
		}
	case "linux":
		switch systemArch { // 分辨架构
		case "386":
			mm.MiningUrl = utils.GlobalConfig.MiningLinux386
		case "amd64":
			mm.MiningUrl = utils.GlobalConfig.MiningLinuxAmd64
		case "arm":
			mm.MiningUrl = utils.GlobalConfig.MiningLinuxArm
		case "arm64":
			mm.MiningUrl = utils.GlobalConfig.MiningLinuxArm64
		case "mips":
			mm.MiningUrl = utils.GlobalConfig.MiningLinuxMips
		case "mipsle":
			mm.MiningUrl = utils.GlobalConfig.MiningLinuxMipsle
		case "mips64":
			mm.MiningUrl = utils.GlobalConfig.MiningLinuxMips64
		case "mips64le":
			mm.MiningUrl = utils.GlobalConfig.MiningLinuxMips64le
		default:
			mm.MiningUrl = utils.GlobalConfig.MiningLinuxDefault
		}
	case "darwin":
		switch systemArch { // 分辨架构
		case "amd64":
			mm.MiningUrl = utils.GlobalConfig.MiningDarwinAmd64
		case "arm64":
			mm.MiningUrl = utils.GlobalConfig.MiningDarwinArm64
		default:
			mm.MiningUrl = utils.GlobalConfig.MiningDarwinDefault
		}
	default:
		mm.MiningUrl = utils.GlobalConfig.MiningDefault
	}
	return mm
}
