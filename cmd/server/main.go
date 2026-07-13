package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	port := getenv("PORT", "8080")
	addr := ":" + port
	// BASE_URL 默认由 PORT 派生，避免只改端口时短链指向错误地址。
	baseURL := getenv("BASE_URL", "http://localhost:"+port)

	// 依赖装配：repository -> service -> handler -> router
	repo := repository.NewMemoryRepository()
	svc := service.NewLinkService(repo)
	h := handler.NewLinkHandler(svc, baseURL)
	r := handler.NewRouter(h)

	// 显式构造 http.Server 并设置超时，防止 Slowloris 等慢速攻击耗尽连接。
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 在独立 goroutine 中启动，主 goroutine 负责监听退出信号做优雅关闭。
	go func() {
		log.Printf("shortlink server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed to start: %v", err)
		}
	}()

	// 等待中断信号（Ctrl+C / kill），优雅关闭：不再接新连接，给在途请求收尾时间。
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server exited")
}
