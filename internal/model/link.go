package model

import "time"

// Link 表示一条短链映射。
// Code 由 ID 经 base62 编码派生（v1 不落库、按需计算），
// 存储层只负责 ID / URL / 点击数 / 创建时间。
type Link struct {
	ID          uint64    `json:"-"`
	Code        string    `json:"code"`
	OriginalURL string    `json:"original_url"`
	ClickCount  uint64    `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
}
