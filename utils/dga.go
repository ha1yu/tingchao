package utils

import (
	"time"
)

// Domain Generation Algorithm 域名固定生成算法
// 根据固定的算法，可以计算出一些固定的域名
// 根据时间戳当种子来计算，算法按照当前月份来计算出10个域名，月份的任意一天这10个域名都一致
// 参考源码如下
// https://github.com/baderj/domain_generation_algorithms/blob/master/locky/dgav3.py

type ConfigDga struct {
	Seed  uint32   // 种子
	Shift uint32   // 偏移量
	TLDs  []string // 域名后缀
}

var configs = []ConfigDga{
	{
		Seed:  7259,
		Shift: 7,
		TLDs: []string{
			"tk", "ml", "ga", "cf", "gq",
		},
	},
}

// -----------------------------------------------------------------------------

func ror32(v, s uint32) uint32 {
	v &= 0xFFFFFFFF
	return ((v >> s) | (v << (32 - s))) & 0xFFFFFFFF
}

func rol32(v, s uint32) uint32 {
	v &= 0xFFFFFFFF
	return ((v << s) | (v >> (32 - s))) & 0xFFFFFFFF
}

// Dga 给定时间戳生成固定的域名
//	date：时间戳
//	configNum 配置文件编号
//	domainNum 种子
func Dga(date time.Time, configNum uint32, domainNum uint32) string {
	c := configs[configNum]
	seed_shifted := rol32(c.Seed, 17)
	dnr_shifted := rol32(domainNum, 21)

	k := uint32(0)
	year := uint32(date.Year())
	month := uint32(date.Month())
	day := uint32(date.Day())

	for i := 0; i < 7; i++ {
		t_0 := ror32(0xB11924E1*(year+k+0x1BF5), c.Shift) & 0xFFFFFFFF
		t_1 := ((t_0 + 0x27100001) ^ k) & 0xFFFFFFFF
		t_2 := (ror32(0xB11924E1*(t_1+c.Seed), c.Shift)) & 0xFFFFFFFF
		t_3 := ((t_2 + 0x27100001) ^ t_1) & 0xFFFFFFFF
		t_4 := (ror32(0xB11924E1*(uint32(day/2)+t_3), c.Shift)) & 0xFFFFFFFF
		t_5 := (0xD8EFFFFF - t_4 + t_3) & 0xFFFFFFFF
		t_6 := (ror32(0xB11924E1*(month+t_5-0x65CAD), c.Shift)) & 0xFFFFFFFF
		t_7 := (t_5 + t_6 + 0x27100001) & 0xFFFFFFFF
		t_8 := (ror32(0xB11924E1*(t_7+seed_shifted+dnr_shifted), c.Shift)) & 0xFFFFFFFF
		//k = ((t_8 + 0x27100001) ^ t_7) & 0xFFFFFFFF	// 原版
		k = ((t_8 + 0x27123456) ^ t_7) & 0xFFFFFFFF // 修改特征
		year++
	}

	length := (k % 11) + 7
	domain := ""
	for i := uint32(0); i < length; i++ {
		k = (ror32(0xB11924E1*rol32(k, i), c.Shift) + 0x27100001) & 0xFFFFFFFF
		domain += string(k%25 + uint32('a'))
	}

	domain += "."

	k = ror32(k*0xB11924E1, c.Shift)
	tlds := c.TLDs
	tld_i := ((k + 0x27100001) & 0xFFFFFFFF) % uint32(len(tlds))
	domain += tlds[tld_i]

	return domain
}

// GetDomain 获取本月的10个域名
// 此方法按照月度来计算域名
// 比如 9月份的每天调用此方法生成的10个域名数组值都是一致的
//
func GetDomain() []string {
	var doaminList []string
	times := get30DayUnixTime()
	for _, time := range times {
		domain := Dga(time, 0, 382)
		doaminList = append(doaminList, domain)
	}
	doaminList = RemoveDuplicatesAndEmpty(doaminList)
	return doaminList
}

// 获取本月的每天中午12的所有时间戳
// 获取每月第1天-20天的值，生成的域名去重后为10个
func get30DayUnixTime() []time.Time {
	var times []time.Time
	now := time.Now()
	for i := 1; i < 21; i++ {
		day01 := now.AddDate(0, 0, -now.Day()+1) // 当前月的第1天
		nextDay := day01.AddDate(0, 0, i)
		nextDayUnix := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 12, 0, 0, 0, now.Location()).Unix()
		times = append(times, time.Unix(nextDayUnix, 0))
	}
	return times
}

// RemoveDuplicatesAndEmpty 去除重复字符串和空格
// 算法生成的域名每2天会有一个重复，去重
func RemoveDuplicatesAndEmpty(a []string) (ret []string) {
	a_len := len(a)
	for i := 0; i < a_len; i++ {
		if (i > 0 && a[i-1] == a[i]) || len(a[i]) == 0 {
			continue
		}
		ret = append(ret, a[i])
	}
	return
}
