module github.com/titan/tingchao

go 1.16


require (
	github.com/gofrs/uuid v4.1.0+incompatible // uuid生成
	github.com/labstack/echo/v4 v4.5.0 // echo框架
	github.com/labstack/gommon v0.3.0 // echo框架依赖
	github.com/spf13/viper v1.9.0 // viper 配置读取
	github.com/fsnotify/fsnotify v1.5.1 // viper依赖的文件监控
)