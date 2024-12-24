package test

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/titan/tingchao/utils"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
)

func InitEcho() {
	e := echo.New()
	e.Use(middleware.Logger()) // 设置日志
	e.GET("/", index)
	e.POST("/", indexPost)
	e.Logger.Fatal(e.Start(":8080"))
}

func index(c echo.Context) error {
	//key1 := c.FormValue("key1")
	//key2 := c.QueryParam("key2")
	//key3 := c.Param("key3")
	//return c.String(http.StatusOK, key1+key2+key3)
	file, err := c.FormFile("txt")
	if err != nil {
		log.Println(err)
		return c.String(http.StatusOK, "错误了！")
	}
	return c.String(http.StatusOK, file.Filename)
}
func indexPost(c echo.Context) error {
	//key1 := c.FormValue("key1")	// 能查到URL参数拼接的
	//key2 := c.QueryParam("key2")	// 能查到URL参数拼接的
	//key3 := c.Param("key3")
	file, err := c.FormFile("txt")
	if err != nil {
		log.Println(err)
		return c.String(http.StatusOK, "错误了！")
	}
	return c.String(http.StatusOK, file.Filename)
}

func tmpPath() {
	path := os.TempDir()
	log.Println(path)
}

// 生成一定长度的随机数
func randomId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil { // 生成一个32位二进制随机数
		return ""
	}
	return base64.URLEncoding.EncodeToString(b) // url编码
}

func test01() {
	for i := 0; i < 100; i++ {
		log.Println(randomId())
	}
}

func test02() {
	// 加解码密钥等于 key=md5(md5("server"+"loveqingqing")+"loveqingqing"+md5("client"+"loveqingqing"))
	key1 := utils.GetMd5Str([]byte("server" + "loveqingqing"))
	key2 := utils.GetMd5Str([]byte("client" + "loveqingqing"))
	key3 := key1 + "loveqingqing" + key2
	key := utils.GetMd5Str([]byte(key3))
	log.Println(key)
}

func TestEcho(t *testing.T) {
	//InitEcho()
	//tmpPath()
	//test01()
	test02()
}
