package main

import (
	"log"
	"os"

	"github.com/dixiaodixiao/shortlink/internal/handler"
	"github.com/dixiaodixiao/shortlink/internal/repository"
	"github.com/dixiaodixiao/shortlink/internal/service"
)

// getenv 读取环境变量，缺省时返回 fallback（12-factor：配置来自环境）。
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	addr := ":" + getenv("PORT", "8080")
	baseURL := getenv("BASE_URL", "http://localhost:8080")

	// 依赖装配：repository -> service -> handler -> router
	repo := repository.NewMemoryRepository()
	svc := service.NewLinkService(repo)
	h := handler.NewLinkHandler(svc, baseURL)
	r := handler.NewRouter(h)

	log.Printf("shortlink server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
