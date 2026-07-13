package repository

import (
	"context"
	"sync"
	"time"

	"github.com/dixiaodixiao/shortlink/internal/model"
)

// MemoryRepository 是 LinkRepository 的内存实现，用于开发与测试，
// 不依赖任何外部数据库。用 RWMutex 保证并发安全。
type MemoryRepository struct {
	mu     sync.RWMutex
	nextID uint64
	byID   map[uint64]*model.Link
}

// 确保 MemoryRepository 满足接口（编译期校验）。
var _ LinkRepository = (*MemoryRepository)(nil)

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		nextID: 1, // 从 1 开始，模拟数据库自增
		byID:   make(map[uint64]*model.Link),
	}
}

func (r *MemoryRepository) Create(ctx context.Context, originalURL string) (*model.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.nextID
	r.nextID++
	link := &model.Link{
		ID:          id,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}
	r.byID[id] = link

	cp := *link // 返回副本，避免调用方直接改到内部状态
	return &cp, nil
}

func (r *MemoryRepository) FindByID(ctx context.Context, id uint64) (*model.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	link, ok := r.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *link
	return &cp, nil
}

func (r *MemoryRepository) IncrementClick(ctx context.Context, id uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	link, ok := r.byID[id]
	if !ok {
		return ErrNotFound
	}
	link.ClickCount++
	return nil
}
