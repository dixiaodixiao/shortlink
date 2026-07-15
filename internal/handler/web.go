package handler

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

// indexHTML 是内嵌的 Web UI 页面。用 embed 打进二进制，
// 部署时无需额外拷贝静态文件，distroless 镜像也能自包含。
//
//go:embed assets/index.html
var indexHTML []byte

// homepage 返回 Web UI 首页。
func homepage(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
}
