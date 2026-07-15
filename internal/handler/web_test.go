package handler

import (
	"net/http"
	"strings"
	"testing"
)

func TestHomepage(t *testing.T) {
	w := doRequest(setupRouter(), http.MethodGet, "/", "")
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, 期望 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, 期望 text/html", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Errorf("首页 body 不是 HTML 文档")
	}
	// 内嵌资源应确实带上了 UI 逻辑，而非空壳
	if !strings.Contains(body, "createLink") {
		t.Errorf("首页缺少生成短链的前端逻辑")
	}
}
