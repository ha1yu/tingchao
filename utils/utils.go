package utils

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"time"
)

var (
	proxyUrl                  = "socks5://127.0.0.1:1080"
	httpTimeOut time.Duration = 10
)

// InitLog logType=1 写入日志文件到 ./log.txt文件，如果 logType=0 则不写入文件
func InitLog(logType int) {
	if logType != 0 {
		//logFile, err := os.OpenFile("./log.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		logFile, err := os.OpenFile("./log_duck_cc_server.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			fmt.Println("open log file failed, err:", err)
			return
		}
		//defer logFile.Close()		//不能关这个写入流，关闭之后就无法持续写入日志
		mw := io.MultiWriter(os.Stdout, logFile) // 通过这种方式可以让你日志信息，同时显示在控制台和写入到log.txt中
		log.SetOutput(mw)
		//log.SetOutput(logFile)
	}
	//log.SetFlags(log.Llongfile | log.Ldate | log.Ltime) // 日志显示日期和时间还有打印日志语句所在行
	log.SetFlags(log.Ldate | log.Ltime) // 日志显示日期和时间
	//log.SetOutput(ioutil.Discard)       // 此设置会将日志直接丢弃,控制台不显示
}

// RandomNumInt 返回一个 a至b 范围内的随机数
func RandomNumInt(a int64, b int64) int64 {
	rang, err := rand.Int(rand.Reader, big.NewInt(b+a-1)) //生成3-7的随机数
	if err != nil {
		return 0
	}
	return rang.Int64() + a
}

func GetMd5Str(b []byte) string {
	return fmt.Sprintf("%x", md5.Sum(b))
}
