package model

import (
	"encoding/gob"

	"github.com/titan/tingchao/utils"
)

// 更新模型

type UpdateModel struct {
	UpdateUrl     string // 软件更新下载地址
	ClientVersion string // 当前最新版本客户端版本号
}

func init() {
	gob.Register(&UpdateModel{})
}

func NewUpdateModel1() *UpdateModel {
	u := &UpdateModel{
		UpdateUrl:     "",
		ClientVersion: "",
	}
	u.ClientVersion = utils.GlobalConfig.ClientVersion // 将最新的客户端版本号下发给客户端
	return u
}

// NewUpdateModel 会根据clientModel中存储的注册信息中系统架构自动判断,应该下载的版本
func NewUpdateModel(cm *ClientModel) *UpdateModel {
	um := &UpdateModel{
		UpdateUrl:     "",
		ClientVersion: "",
	}
	um.ClientVersion = utils.GlobalConfig.ClientVersion // 将最新的客户端版本号下发给客户端

	systemType := cm.ClientRegModel.SystemType
	systemArch := cm.ClientRegModel.SystemArch
	switch systemType { // 不同的系统不同架构下载不同的更新包
	case "windows":
		switch systemArch { // 分辨架构
		case "386":
			um.UpdateUrl = utils.GlobalConfig.UpdateWindows386
		case "amd64":
			um.UpdateUrl = utils.GlobalConfig.UpdateWindowsAmd64
		default:
			um.UpdateUrl = utils.GlobalConfig.UpdateWindowsDefault
		}
	case "linux":
		switch systemArch { // 分辨架构
		case "386":
			um.UpdateUrl = utils.GlobalConfig.UpdateLinux386
		case "amd64":
			um.UpdateUrl = utils.GlobalConfig.UpdateLinuxAmd64
		case "arm":
			um.UpdateUrl = utils.GlobalConfig.UpdateLinuxArm
		case "arm64":
			um.UpdateUrl = utils.GlobalConfig.UpdateLinuxArm64
		case "mips":
			um.UpdateUrl = utils.GlobalConfig.UpdateLinuxMips
		case "mipsle":
			um.UpdateUrl = utils.GlobalConfig.UpdateLinuxMipsle
		case "mips64":
			um.UpdateUrl = utils.GlobalConfig.UpdateLinuxMips64
		case "mips64le":
			um.UpdateUrl = utils.GlobalConfig.UpdateLinuxMips64le
		default:
			um.UpdateUrl = utils.GlobalConfig.UpdateLinuxDefault
		}
	case "darwin":
		switch systemArch { // 分辨架构
		case "amd64":
			um.UpdateUrl = utils.GlobalConfig.UpdateDarwinAmd64
		case "arm64":
			um.UpdateUrl = utils.GlobalConfig.UpdateDarwinArm64
		default:
			um.UpdateUrl = utils.GlobalConfig.UpdateDarwinDefault
		}
	default:
		um.UpdateUrl = utils.GlobalConfig.UpdateDefault
	}
	return um
}
