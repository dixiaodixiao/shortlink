package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dixiaodixiao/shortlink/internal/repository"
	"github.com/dixiaodixiao/shortlink/internal/service"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	repo := repository.NewMemoryRepository()
	svc := service.NewLinkService(repo)
	h := NewLinkHandler(svc, "http://localhost:8080")
	return NewRouter(h)
}

// 发起请求的小工具
func doRequest(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHealthz(t *testing.T) {
	w := doRequest(setupRouter(), http.MethodGet, "/healthz", "")
	if w.Code != http.StatusOK {
		t.Errorf("状态码 = %d, 期望 200", w.Code)
	}
}

func TestCreateLink(t *testing.T) {
	w := doRequest(setupRouter(), http.MethodPost, "/api/links", `{"url":"https://example.com"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("状态码 = %d, 期望 201, body=%s", w.Code, w.Body.String())
	}
	var resp createResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if resp.Code == "" || resp.ShortURL == "" {
		t.Errorf("响应缺少 code 或 short_url: %+v", resp)
	}
}

func TestCreateLink_InvalidURL(t *testing.T) {
	w := doRequest(setupRouter(), http.MethodPost, "/api/links", `{"url":"not-a-url"}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("状态码 = %d, 期望 400", w.Code)
	}
}

func TestCreateLink_BadJSON(t *testing.T) {
	w := doRequest(setupRouter(), http.MethodPost, "/api/links", `{bad json`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("状态码 = %d, 期望 400", w.Code)
	}
}

func TestRedirect(t *testing.T) {
	r := setupRouter()
	// 先创建
	cw := doRequest(r, http.MethodPost, "/api/links", `{"url":"https://example.com/target"}`)
	var resp createResponse
	json.Unmarshal(cw.Body.Bytes(), &resp)

	// 再访问短码
	w := doRequest(r, http.MethodGet, "/"+resp.Code, "")
	if w.Code != http.StatusFound {
		t.Fatalf("状态码 = %d, 期望 302", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "https://example.com/target" {
		t.Errorf("Location = %q, 期望 https://example.com/target", loc)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	w := doRequest(setupRouter(), http.MethodGet, "/zzz", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("状态码 = %d, 期望 404", w.Code)
	}
}

func TestDetail_ShowsClickCount(t *testing.T) {
	r := setupRouter()
	cw := doRequest(r, http.MethodPost, "/api/links", `{"url":"https://example.com"}`)
	var resp createResponse
	json.Unmarshal(cw.Body.Bytes(), &resp)

	// 访问两次，点击数应为 2
	doRequest(r, http.MethodGet, "/"+resp.Code, "")
	doRequest(r, http.MethodGet, "/"+resp.Code, "")

	w := doRequest(r, http.MethodGet, "/api/links/"+resp.Code, "")
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, 期望 200", w.Code)
	}
	var detail struct {
		ClickCount uint64 `json:"click_count"`
	}
	json.Unmarshal(w.Body.Bytes(), &detail)
	if detail.ClickCount != 2 {
		t.Errorf("点击数 = %d, 期望 2", detail.ClickCount)
	}
}
