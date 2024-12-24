package httpmodel

type ClientCountModel struct {
	Code int16  // json状态码
	Msg  string // json消息

	ClientCount       int // session中的客户端总数
	ClientCount10Min  int // 10分钟内活跃的客户端总数
	ClientCount1Hour  int // 1小时内活跃的客户端总数
	ClientCount24Hour int // 24小时内活跃的客户端总数
	LinuxClientCount  int // 连接的linux客户端总数
	WinClientCount    int // 连接的win客户端总数
	OtherClientCount  int // 连接的其他客户端总数
}
