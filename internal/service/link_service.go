package service

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/dixiaodixiao/shortlink/internal/model"
	"github.com/dixiaodixiao/shortlink/internal/repository"
)

// ErrInvalidURL 表示传入的原始链接不合法。
var ErrInvalidURL = errors.New("service: 非法的 URL")

// LinkService 承载短链业务逻辑，只依赖 repository 接口。
type LinkService struct {
	repo repository.LinkRepository
}

func NewLinkService(repo repository.LinkRepository) *LinkService {
	return &LinkService{repo: repo}
}

// Create 校验并保存原始链接，返回带短码的记录。
// 短码 = base62(自增 ID)，体现"数据库发号 + 编码"的思路。
func (s *LinkService) Create(ctx context.Context, originalURL string) (*model.Link, error) {
	originalURL = strings.TrimSpace(originalURL)
	if err := validateURL(originalURL); err != nil {
		return nil, err
	}
	link, err := s.repo.Create(ctx, originalURL)
	if err != nil {
		return nil, err
	}
	link.Code = EncodeBase62(link.ID)
	return link, nil
}

// Resolve 用于重定向：按短码找到原始链接，并累加点击数。
// 非法短码统一按"不存在"处理，避免向外暴露内部编码细节。
func (s *LinkService) Resolve(ctx context.Context, code string) (*model.Link, error) {
	id, err := DecodeBase62(code)
	if err != nil {
		return nil, repository.ErrNotFound
	}
	link, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	link.Code = code
	// 点击计数：尽力而为，失败不影响重定向本身
	_ = s.repo.IncrementClick(ctx, id)
	return link, nil
}

// GetByCode 用于查详情：只读，不累加点击数。
func (s *LinkService) GetByCode(ctx context.Context, code string) (*model.Link, error) {
	id, err := DecodeBase62(code)
	if err != nil {
		return nil, repository.ErrNotFound
	}
	link, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	link.Code = code
	return link, nil
}

// validateURL 只接受 http/https 且带主机名的链接，
// 防止存入 javascript:、file:// 等危险协议（安全考虑）。
func validateURL(raw string) error {
	if raw == "" {
		return ErrInvalidURL
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ErrInvalidURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrInvalidURL
	}
	if u.Host == "" {
		return ErrInvalidURL
	}
	return nil
}
