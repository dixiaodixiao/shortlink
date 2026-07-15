package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter 装配全部路由。main 与测试共用，保证测的就是线上跑的路由。
func NewRouter(h *LinkHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Web UI 首页
	r.GET("/", homepage)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		api.POST("/links", h.Create)
		api.GET("/links/:code", h.Detail)
	}

	// 根路径短码重定向，注册在最后
	r.GET("/:code", h.Redirect)

	return r
}
