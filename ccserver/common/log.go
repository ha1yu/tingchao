package common

import (
	"io"
	"log"
	"os"
)

func init() {
	initLog(Configs.Debug)
}

func initLog(debug bool) {
	logFile, err := os.OpenFile("./tingchao-server.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Println("日志文件打开错误:", err)
		return
	}
	//不能关这个写入流,关闭之后就无法持续写入日志
	//defer logFile.Close()

	// 通过这种方式可以让你日志信息,同时显示在控制台和写入到log.txt中
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	if debug {
		log.SetFlags(log.Llongfile | log.Ldate | log.Ltime) // 日志显示日期和时间还有打印日志语句所在行
	} else {
		log.SetFlags(log.Ldate | log.Ltime) // 日志显示日期和时间,不显示行号
	}

	//log.SetFlags(log.Llongfile | log.Ldate | log.Ltime) // 日志显示日期和时间还有打印日志语句所在行

	//log.SetOutput(ioutil.Discard) // 此设置会将日志直接丢弃,控制台不显示
}
