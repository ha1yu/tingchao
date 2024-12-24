package test

import (
	"github.com/titan/tingchao/utils"
	"log"
	"testing"
)

func Test(t *testing.T) {
	//InitEcho()
	//tmpPath()
	//test01()
	type1 := utils.DownloadFile("http://hcfuse.com/static/js/tools/c/linux64", "/Users/admin/Downloads", "abc123")
	if type1 {
		log.Println("ok")
	} else {
		log.Println("no ok")
	}
}
