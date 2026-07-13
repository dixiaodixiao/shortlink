package service

import "fmt"

// base62 字符表：0-9、a-z、A-Z，共 62 个字符。
// 顺序决定了编码结果，一旦上线不可随意更改（否则旧短码会解析错误）。
const base62Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const base = 62

// EncodeBase62 将无符号整数（通常是数据库自增 ID）编码为 base62 短码。
// id 为 0 时返回 "0"。
func EncodeBase62(id uint64) string {
	if id == 0 {
		return string(base62Alphabet[0])
	}

	// 逐位取余，得到的是低位在前，最后反转。
	var buf []byte
	for id > 0 {
		buf = append(buf, base62Alphabet[id%base])
		id /= base
	}
	// 反转
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

// decodeMap 是字符 -> 数值的反查表，初始化一次，解码时 O(1) 查找。
var decodeMap = func() map[byte]uint64 {
	m := make(map[byte]uint64, base)
	for i := 0; i < len(base62Alphabet); i++ {
		m[base62Alphabet[i]] = uint64(i)
	}
	return m
}()

// DecodeBase62 将 base62 短码还原为整数。
// 遇到不在字符表中的字符返回错误。
func DecodeBase62(code string) (uint64, error) {
	if code == "" {
		return 0, fmt.Errorf("base62: 短码为空")
	}
	const maxUint64 = ^uint64(0)
	var n uint64
	for i := 0; i < len(code); i++ {
		v, ok := decodeMap[code[i]]
		if !ok {
			return 0, fmt.Errorf("base62: 非法字符 %q", code[i])
		}
		// 溢出检测：若 n*base+v 会超过 uint64 上限则拒绝，
		// 避免静默回绕导致解析到错误的 ID。
		if n > (maxUint64-v)/base {
			return 0, fmt.Errorf("base62: 短码超出取值范围")
		}
		n = n*base + v
	}
	return n, nil
}
