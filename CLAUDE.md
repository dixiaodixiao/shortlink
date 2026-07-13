# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目简介

一个用 Go + Gin 实现的**短链接服务**。领域简单，但在编码/发号、分层、可测试性、可观测性上做到生产级思路。完整设计见 `DESIGN.md`。

## 常用命令

```bash
go build ./...                                          # 构建
go test ./...                                           # 跑全部测试
go test ./internal/service/ -run TestEncodeBase62 -v    # 跑单个测试
go vet ./...                                            # 静态检查
go run ./cmd/server                                     # 启动服务（默认 :8080）
PORT=8081 BASE_URL=http://localhost:8081 go run ./cmd/server  # 指定端口启动
```

- **`go test -race` 需要 cgo（本机无 gcc，跑不了）**，竞争检测放到 CI（Linux）里执行。

## 架构

严格三层，依赖单向向下，每层可独立测试（对下层 mock 或用内存实现）：

```
handler (HTTP/Gin)  →  service (业务逻辑)  →  repository (存储)
```

- **依赖装配集中在 `cmd/server/main.go`**：`repo → service → handler → router` 手动注入。没有 DI 框架，加依赖就在这里串。
- **存储面向接口**：`repository.LinkRepository` 是抽象，当前只有 `MemoryRepository` 内存实现（`var _ LinkRepository = (*MemoryRepository)(nil)` 做编译期校验）。Postgres/Redis 实现待加，加时只需满足该接口、在 main 里替换，上层无需改动。
- **路由共用**：`handler.NewRouter` 同时被 main 和测试使用，保证测试测的就是线上路由。根路径通配 `/:code`（重定向）与静态路由（`/healthz`、`/api/...`）共存，通配路由**注册在最后**。

## 关键不变量（改动前务必理解）

- **短码 = base62(数据库自增 ID)，且不落库、按需计算**。`service.Create` 拿到自增 ID 后用 `EncodeBase62` 得到短码；`Resolve`/`GetByCode` 反向 `DecodeBase62` 得到 ID。
- **短码必须规范化**：一个链接只能有一个有效短码。解析统一走 `LinkService.codeToID`，它会校验 `EncodeBase62(id) == code`，借此拒绝前导零等非规范写法（如 `/01` 不等价于 `/1`，返回 404）。新增任何按短码查询的路径都应走 `codeToID`，不要直接用 `DecodeBase62`。
- **`DecodeBase62` 有 uint64 溢出保护**：超长短码会返回错误而非静默回绕。

## 配置与运行

- 配置来自环境变量（12-factor）：`PORT`、`BASE_URL`、`GIN_MODE`。`BASE_URL` 未设时默认由 `PORT` 派生，避免只改端口导致返回的 `short_url` 指向错误地址。
- 服务用显式 `http.Server`（设了读写/空闲超时）+ 监听 SIGINT/SIGTERM 做优雅关闭。

## 本机开发环境注意

- 本机配了**全局 HTTP 代理**：直接用 `curl` / `curl.exe` 访问 `localhost` 会返回 **502**。测本地服务请用 PowerShell 的 `Invoke-RestMethod`（对 localhost 自动绕代理），或 .NET `HttpClient` 设 `UseProxy=$false`（`AllowAutoRedirect=$false` 可查看 302 响应头）。

## 提交规范

使用 Conventional Commits：`feat:` / `fix:` / `chore:` / `docs:` 等前缀。
