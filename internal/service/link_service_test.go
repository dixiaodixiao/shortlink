package service

import (
	"context"
	"errors"
	"testing"

	"github.com/dixiaodixiao/shortlink/internal/repository"
)

func newTestService() *LinkService {
	return NewLinkService(repository.NewMemoryRepository())
}

func TestLinkService_Create(t *testing.T) {
	svc := newTestService()
	link, err := svc.Create(context.Background(), "https://example.com/path")
	if err != nil {
		t.Fatalf("Create 出错: %v", err)
	}
	if link.Code == "" {
		t.Error("短码不应为空")
	}
	// 首条记录 ID=1，base62(1)="1"
	if link.Code != "1" {
		t.Errorf("首条短码 = %q, 期望 \"1\"", link.Code)
	}
}

func TestLinkService_Create_InvalidURL(t *testing.T) {
	svc := newTestService()
	cases := []string{"", "   ", "not-a-url", "ftp://x.com", "javascript:alert(1)", "http://"}
	for _, c := range cases {
		if _, err := svc.Create(context.Background(), c); !errors.Is(err, ErrInvalidURL) {
			t.Errorf("Create(%q) 期望 ErrInvalidURL, 得到 %v", c, err)
		}
	}
}

func TestLinkService_Resolve(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	created, _ := svc.Create(ctx, "https://example.com")

	got, err := svc.Resolve(ctx, created.Code)
	if err != nil {
		t.Fatalf("Resolve 出错: %v", err)
	}
	if got.OriginalURL != "https://example.com" {
		t.Errorf("URL = %q, 期望 https://example.com", got.OriginalURL)
	}
}

func TestLinkService_Resolve_NotFound(t *testing.T) {
	svc := newTestService()
	// "zzz" 是合法 base62，但对应 ID 不存在
	if _, err := svc.Resolve(context.Background(), "zzz"); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("期望 ErrNotFound, 得到 %v", err)
	}
}

func TestLinkService_Resolve_InvalidCode(t *testing.T) {
	svc := newTestService()
	// "!!!" 非法字符，也应按不存在处理
	if _, err := svc.Resolve(context.Background(), "!!!"); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("非法短码期望 ErrNotFound, 得到 %v", err)
	}
}

func TestLinkService_Resolve_IncrementsClick(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	created, _ := svc.Create(ctx, "https://example.com")

	svc.Resolve(ctx, created.Code)
	svc.Resolve(ctx, created.Code)

	detail, _ := svc.GetByCode(ctx, created.Code)
	if detail.ClickCount != 2 {
		t.Errorf("点击数 = %d, 期望 2", detail.ClickCount)
	}
}

func TestLinkService_GetByCode_NoIncrement(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	created, _ := svc.Create(ctx, "https://example.com")

	svc.GetByCode(ctx, created.Code)
	detail, _ := svc.GetByCode(ctx, created.Code)
	if detail.ClickCount != 0 {
		t.Errorf("查详情不应增加点击数, 得到 %d", detail.ClickCount)
	}
}
