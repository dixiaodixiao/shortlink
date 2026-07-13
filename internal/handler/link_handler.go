package handler

import (
	"errors"
	"net/http"

	"github.com/dixiaodixiao/shortlink/internal/repository"
	"github.com/dixiaodixiao/shortlink/internal/service"
	"github.com/gin-gonic/gin"
)

// LinkHandler 处理短链相关的 HTTP 请求。
type LinkHandler struct {
	svc     *service.LinkService
	baseURL string // 用于拼接返回给用户的完整短链，如 http://localhost:8080
}

func NewLinkHandler(svc *service.LinkService, baseURL string) *LinkHandler {
	return &LinkHandler{svc: svc, baseURL: baseURL}
}

type createRequest struct {
	URL string `json:"url"`
}

type createResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

// Create 处理 POST /api/links
func (h *LinkHandler) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体必须是 JSON，且包含 url 字段"})
		return
	}

	link, err := h.svc.Create(c.Request.Context(), req.URL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "非法的 URL"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}

	c.JSON(http.StatusCreated, createResponse{
		Code:     link.Code,
		ShortURL: h.baseURL + "/" + link.Code,
	})
}

// Redirect 处理 GET /:code，302 跳转到原始链接
func (h *LinkHandler) Redirect(c *gin.Context) {
	link, err := h.svc.Resolve(c.Request.Context(), c.Param("code"))
	if err != nil {
		h.writeLookupError(c, err)
		return
	}
	c.Redirect(http.StatusFound, link.OriginalURL)
}

// Detail 处理 GET /api/links/:code，返回短链详情（含点击数）
func (h *LinkHandler) Detail(c *gin.Context) {
	link, err := h.svc.GetByCode(c.Request.Context(), c.Param("code"))
	if err != nil {
		h.writeLookupError(c, err)
		return
	}
	c.JSON(http.StatusOK, link)
}

// writeLookupError 统一处理"查不到"类错误的响应。
func (h *LinkHandler) writeLookupError(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "短链不存在"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
}
