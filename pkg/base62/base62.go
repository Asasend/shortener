package base62

import (
	"math"
	"strings"
)

// 62进制转换的模块

// 0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
// 打乱：
// kP0rE7GqLmZj8dWF3tNs5iUxCvY4bKu9aQoHI6fB1yMhRg2DOeSzXjTN

// 0-9 : 0-9
// a-z: 10-35
// A-Z: 36-61

// 如和实现62进制转换

// 为了避免被人恶意请求，我们可以打乱这个字符串
// const base62str = `kP0rE7GqLmZj8dWF3tNs5iUxCvY4bKu9aQoHI6fB1yMhRg2DOeSzXjTN`

var (
	base62str string
)

// MustInit 要使用base62这包必须要调用该函数完成初始化
func MustInit(bs string) {
	if len(bs) == 0 {
		panic("need base string!")
	}
	base62str = bs
}

func Int2String(seq uint64) string {
	if seq == 0 {
		return string(base62str[0])
	}

	bl := []byte{}

	for seq > 0 {
		mod := seq % 62
		div := seq / 62
		bl = append(bl, []byte(base62str)[mod])
		seq = div
	}
	// 反转
	reverse(bl)
	return string(bl)
}

// reverseBytes 反转字节切片
func reverse(bl []byte) {
	for i := 0; i < len(bl)/2; i++ {
		bl[i], bl[len(bl)-1-i] = bl[len(bl)-1-i], bl[i]
	}
}

// String2Int 62进制字符串转为10进制数

func String2Int(s string) (seq uint64) {
	bl := []byte(s)
	reverse(bl)
	for idx, b := range bl {
		base := math.Pow(62, float64(idx))
		seq += uint64(strings.Index(base62str, string(b))) * uint64(base)
	}
	return seq
}
