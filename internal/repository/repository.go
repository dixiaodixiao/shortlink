package repository

import (
	"context"
	"errors"

	"github.com/dixiaodixiao/shortlink/internal/model"
)

// ErrNotFound 表示查询的记录不存在。上层据此返回 404。
var ErrNotFound = errors.New("repository: 记录不存在")

// LinkRepository 是短链存储的抽象。
// 内存版、Postgres 版都实现这个接口，上层（service）只依赖接口，
// 这样底层存储可无缝替换（依赖倒置）。
type LinkRepository interface {
	// Create 保存原始链接，返回带有自增 ID 的记录。
	Create(ctx context.Context, originalURL string) (*model.Link, error)
	// FindByID 按 ID 查询；不存在返回 ErrNotFound。
	FindByID(ctx context.Context, id uint64) (*model.Link, error)
	// IncrementClick 将指定记录的点击数 +1。
	IncrementClick(ctx context.Context, id uint64) error
}
