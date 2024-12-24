package model

import (
	"encoding/json"
	"github.com/titan/tingchao/ccserver/common"
	"github.com/titan/tingchao/utils"
	"log"
)

// 注册模型

type RegModel struct {
	Id            string `gorm:"column:id;primary_key;type:varchar(36)" json:"Id"`             // 客户端ID
	ClientVersion string `gorm:"column:client_version;type:varchar(255)" json:"ClientVersion"` // 当前client版本号
	Uid           string `gorm:"column:uid;type:varchar(255)" json:"Uid"`                      // 用户ID
	Gid           string `gorm:"column:gid;type:varchar(255)" json:"Gid"`                      // 初级组ID
	Username      string `gorm:"column:user_name;type:varchar(255)" json:"Username"`           // 用户名
	Name          string `gorm:"column:name;type:varchar(255)" json:"Name"`                    // 用户名字
	HomeDir       string `gorm:"column:home_dir;type:varchar(255)" json:"HomeDir"`             // 用户文件夹
	SystemType    string `gorm:"column:system_type;type:varchar(255)" json:"SystemType"`       // 系统类型 windows linux darwin
	SystemArch    string `gorm:"column:system_arch;type:varchar(255)" json:"SystemArch"`       // 系统架构 386 amd64 arm ppc64
	Hostname      string `gorm:"column:host_name;type:varchar(255)" json:"Hostname"`           // hostname

	NetworkList []NetWorkModel `json:"network"`
}

type NetWorkModel struct {
	Ethernet   string `json:"ethernet"`   // 网卡
	IpAddress  string `json:"ipAddress"`  // IP
	MacAddress string `json:"macAddress"` // MAC
	Mask       string `json:"mask"`
}

var RegModelDao *RegModel

func init() {
	RegModelDao = newRegModelDao()
}

// TableName 设置表名
func (r RegModel) TableName() string {
	return "reg"
}

func NewRegModel() *RegModel {
	rm := &RegModel{}
	return rm
}

func newRegModelDao() *RegModel {
	rm := &RegModel{}
	return rm
}

func GetClientRegCode(rmStr string) string {
	rm := NewRegModel()

	err := json.Unmarshal([]byte(rmStr), rm)
	if err != nil {
		log.Println(err)
		return ""
	}

	str := rm.SystemType + rm.SystemArch + rm.Hostname
	networkList := rm.NetworkList

	for _, netWorkModel := range networkList {
		str = str + netWorkModel.Ethernet
		str = str + netWorkModel.IpAddress
		str = str + netWorkModel.MacAddress
	}

	str = str + "client_salt"
	id := utils.GetMd5Str([]byte(str))
	return id
}

func (r RegModel) String() string {
	str, err := json.Marshal(r)
	if err != nil {
		log.Println(err)
		return ""
	}
	return string(str)
}

func (r RegModel) FindAll() ([]RegModel, error) {
	var rm []RegModel
	err := common.Db.Table(r.TableName()).Find(&rm).Error
	return rm, err
}

func (r RegModel) FindByID(id string) (RegModel, error) {
	var rm RegModel
	err := common.Db.Table(r.TableName()).Where("id=?", id).Find(&rm).Error
	return rm, err
}

func (r RegModel) AddOne(rm RegModel) error {
	err := common.Db.Create(rm).Error
	return err
}

func (r RegModel) AddList(rms []RegModel) error {
	// 开启事务
	tx := common.Db.Begin()

	for rm := range rms {
		err := tx.Create(rm).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}
