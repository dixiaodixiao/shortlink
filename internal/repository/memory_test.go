package repository

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestMemoryRepository_CreateAndFind(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	link, err := repo.Create(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Create 出错: %v", err)
	}
	if link.ID != 1 {
		t.Errorf("首条记录 ID = %d, 期望 1", link.ID)
	}

	got, err := repo.FindByID(ctx, link.ID)
	if err != nil {
		t.Fatalf("FindByID 出错: %v", err)
	}
	if got.OriginalURL != "https://example.com" {
		t.Errorf("URL = %q, 期望 https://example.com", got.OriginalURL)
	}
}

func TestMemoryRepository_AutoIncrement(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	a, _ := repo.Create(ctx, "https://a.com")
	b, _ := repo.Create(ctx, "https://b.com")
	if b.ID != a.ID+1 {
		t.Errorf("ID 未自增: a=%d b=%d", a.ID, b.ID)
	}
}

func TestMemoryRepository_NotFound(t *testing.T) {
	repo := NewMemoryRepository()
	_, err := repo.FindByID(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("期望 ErrNotFound, 得到 %v", err)
	}
}

func TestMemoryRepository_IncrementClick(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()
	link, _ := repo.Create(ctx, "https://example.com")

	for i := 0; i < 3; i++ {
		if err := repo.IncrementClick(ctx, link.ID); err != nil {
			t.Fatalf("IncrementClick 出错: %v", err)
		}
	}
	got, _ := repo.FindByID(ctx, link.ID)
	if got.ClickCount != 3 {
		t.Errorf("ClickCount = %d, 期望 3", got.ClickCount)
	}
}

// 并发写入，配合 go test -race 检测数据竞争
func TestMemoryRepository_Concurrent(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			repo.Create(ctx, "https://example.com")
		}()
	}
	wg.Wait()

	// 100 次并发 Create 后，下一个 ID 应为 101
	link, _ := repo.Create(ctx, "https://example.com")
	if link.ID != 101 {
		t.Errorf("并发后 ID = %d, 期望 101（说明自增有竞争）", link.ID)
	}
}
