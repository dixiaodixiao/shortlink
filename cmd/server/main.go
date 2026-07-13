package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 健康检查：用于确认服务存活（容器编排、负载均衡都会用到）
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	const addr = ":8080"
	log.Printf("shortlink server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
