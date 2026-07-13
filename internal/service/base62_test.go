package service

import "testing"

func TestEncodeBase62(t *testing.T) {
	tests := []struct {
		name string
		id   uint64
		want string
	}{
		{"零", 0, "0"},
		{"个位", 1, "1"},
		{"数字边界9", 9, "9"},
		{"小写起点10", 10, "a"},
		{"大写起点36", 36, "A"},
		{"进位边界62", 62, "10"},
		{"较大值", 12345, "3d7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeBase62(tt.id); got != tt.want {
				t.Errorf("EncodeBase62(%d) = %q, 期望 %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestDecodeBase62(t *testing.T) {
	tests := []struct {
		name string
		code string
		want uint64
	}{
		{"零", "0", 0},
		{"个位", "1", 1},
		{"小写", "a", 10},
		{"大写", "A", 36},
		{"进位", "10", 62},
		{"较大值", "3d7", 12345},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeBase62(tt.code)
			if err != nil {
				t.Fatalf("DecodeBase62(%q) 返回错误: %v", tt.code, err)
			}
			if got != tt.want {
				t.Errorf("DecodeBase62(%q) = %d, 期望 %d", tt.code, got, tt.want)
			}
		})
	}
}

// 非法字符应返回错误
func TestDecodeBase62_Invalid(t *testing.T) {
	if _, err := DecodeBase62("!@#"); err == nil {
		t.Error("非法短码应返回错误，但 err 为 nil")
	}
}

// 编码后再解码应还原为原值（往返一致性，property 测试思路）
func TestEncodeDecodeRoundTrip(t *testing.T) {
	for _, id := range []uint64{0, 1, 61, 62, 100, 999999, 1 << 40} {
		code := EncodeBase62(id)
		got, err := DecodeBase62(code)
		if err != nil {
			t.Fatalf("往返解码 %q 出错: %v", code, err)
		}
		if got != id {
			t.Errorf("往返不一致: id=%d -> code=%q -> %d", id, code, got)
		}
	}
}
