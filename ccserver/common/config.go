package common

import (
	"encoding/json"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"strings"
)

var Configs *Config

func init() {
	InitConfig()
}

type Config struct {
	Debug         bool
	Name          string // 当前服务器名称
	CcServerPort  string // 当前服务器主机监听端口号
	WebServerPort string // 当前服务器主机监听端口号
	ServerVersion string // 当前服务端版本号
	MessageAesKey string // AES加密密钥,服务端客户端需要一致,服务端为了性能直接写死,客户端不写死,每次计算出来
	MsgHistoryLen int    // 历史CC消息数组长度，默认记录5条命令
	ClientVersion string // 客户端版本号

	Mysql   *MysqlConfig
	Redis   *RedisConfig
	Session *SessionConfig
	Update  *UpdateConfig
}

type MysqlConfig struct {
	Hostname string // 数据库ip地址
	Username string // 数据库用户名
	Password string // 数据库密码
	Port     string // 数据库端口
	DbName   string // 数据库名
}

type RedisConfig struct {
	Hostname string // redis ip地址
	Port     string // redis 端口
}

type SessionConfig struct {
	SessionTime       int    // Session有效期,单位是天
	SessionGcTime     int    // session GC 的时间,默认是1天
	SessionCookieName string // session对应的cookie名称,默认JSESSIONID
}

type UpdateConfig struct {
	UpdateWindows386     string
	UpdateWindowsAmd64   string
	UpdateWindowsDefault string
	UpdateLinux386       string
	UpdateLinuxAmd64     string
	UpdateLinuxArm       string
	UpdateLinuxArm64     string
	UpdateLinuxMips      string
	UpdateLinuxMipsle    string
	UpdateLinuxMips64    string
	UpdateLinuxMips64le  string
	UpdateLinuxDefault   string
	UpdateDarwinAmd64    string
	UpdateDarwinArm64    string
	UpdateDarwinDefault  string
	UpdateDefault        string
}

func InitConfig() {
	viper.SetConfigName("config")                          // 配置文件前缀
	viper.SetConfigType("yml")                             // 配置文件后缀
	viper.AddConfigPath("./")                              // 绑定配置路径
	viper.AutomaticEnv()                                   // 绑定全部环境变量
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // 字符串替换

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("读取配置文件出错", err)
	}
	viper.WatchConfig() // 开启热加载，如果配置文件发生改变则热加载

	viper.OnConfigChange(func(e fsnotify.Event) { // 热加载的回调
		Configs = GetConfig()
		log.Println("==> 重新加载配置文件", e.Name, e.Op)
	})

	Configs = GetConfig()
}

func GetConfig() *Config {
	var config = &Config{
		Debug:         viper.GetBool("Debug"),
		Name:          viper.GetString("name"),
		CcServerPort:  viper.GetString("CcServerPort"),
		ServerVersion: viper.GetString("ServerVersion"),
		MessageAesKey: viper.GetString("MessageAesKey"),
		MsgHistoryLen: viper.GetInt("MsgHistoryLen"),
		ClientVersion: viper.GetString("ClientVersion"),
		Mysql: &MysqlConfig{
			Hostname: viper.GetString("mysql.Hostname"),
			Username: viper.GetString("mysql.Username"),
			Password: viper.GetString("mysql.Password"),
			Port:     viper.GetString("mysql.Port"),
			DbName:   viper.GetString("mysql.DbName"),
		},
		Redis: &RedisConfig{
			Hostname: viper.GetString("redis.Hostname"),
			Port:     viper.GetString("redis.Port"),
		},
		Session: &SessionConfig{
			SessionTime:       viper.GetInt("Session.SessionTime"),
			SessionGcTime:     viper.GetInt("Session.SessionGcTime"),
			SessionCookieName: viper.GetString("Session.SessionCookieName"),
		},
		Update: &UpdateConfig{
			UpdateWindows386:     viper.GetString("update.UpdateWindows386"),
			UpdateWindowsAmd64:   viper.GetString("update.UpdateWindowsAmd64"),
			UpdateWindowsDefault: viper.GetString("update.UpdateWindowsDefault"),
			UpdateLinux386:       viper.GetString("update.UpdateLinux386"),
			UpdateLinuxAmd64:     viper.GetString("update.UpdateLinuxAmd64"),
			UpdateLinuxArm:       viper.GetString("update.UpdateLinuxArm"),
			UpdateLinuxArm64:     viper.GetString("update.UpdateLinuxArm64"),
			UpdateLinuxMips:      viper.GetString("update.UpdateLinuxMips"),
			UpdateLinuxMipsle:    viper.GetString("update.UpdateLinuxMipsle"),
			UpdateLinuxMips64:    viper.GetString("update.UpdateLinuxMips64"),
			UpdateLinuxMips64le:  viper.GetString("update.UpdateLinuxMips64le"),
			UpdateLinuxDefault:   viper.GetString("update.UpdateLinuxDefault"),
			UpdateDarwinAmd64:    viper.GetString("update.UpdateDarwinAmd64"),
			UpdateDarwinArm64:    viper.GetString("update.UpdateDarwinArm64"),
			UpdateDarwinDefault:  viper.GetString("update.UpdateDarwinDefault"),
			UpdateDefault:        viper.GetString("update.UpdateDefault"),
		},
	}
	return config
}

func (c Config) String() string {
	str, _ := json.Marshal(c)
	return string(str)
}
