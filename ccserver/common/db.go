package common

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var Db *gorm.DB

func init() {
	//Db = InitDb()
}

func InitDb() *gorm.DB {

	var err error
	var db *gorm.DB

	dbConnConf := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=60s",
		Configs.Mysql.Username,
		Configs.Mysql.Password,
		Configs.Mysql.Hostname,
		Configs.Mysql.Port,
		Configs.Mysql.DbName,
	)
	db, err = gorm.Open(mysql.Open(dbConnConf), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})

	if err != nil {
		log.Errorf("连接数据库异常: %v", err.Error())
		panic(err)
	}

	sqlDb, _ := db.DB()
	sqlDb.SetMaxIdleConns(50) //设置最大连接数
	sqlDb.SetMaxOpenConns(10) //设置最大的空闲连接数
	//data, _ := json.Marshal(sqlDb.Stats()) //获得当前的SQL配置情况
	//log.Printf(string(data))
	return db
}
