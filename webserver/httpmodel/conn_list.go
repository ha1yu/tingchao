package httpmodel

type ConnList struct {
	Code       int16    // json状态吗
	ClientList []Client // 客户端对象
	Msg        string   // json消息
}

type Client struct {
	SessionID        string //当前连接的ID 也可以称作为SessionID，ID全局唯一
	IPAddress        string //客户端地址
	LastAccessedTime int64  // 最后访问时间
	Version          string //当前duck-cc-client版本号

	Uid      string // 用户ID
	Gid      string // 用户组ID
	Username string // 用户名
	Name     string // 用户名字
	HomeDir  string // 用户文件夹

	SystemType string //系统类型 windows linux darwin
	SystemArch string //系统架构	386 amd64 arm ppc64

	Hostname string // hostname

}
