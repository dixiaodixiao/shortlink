# syntax=docker/dockerfile:1

# ---- 构建阶段 ----
FROM golang:1.26-alpine AS builder
WORKDIR /src

# 先只拷贝依赖清单，命中缓存，依赖不变时不重复下载
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# 静态编译（CGO_ENABLED=0），去符号表减小体积
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

# ---- 运行阶段 ----
# distroless static：无 shell、无包管理器，攻击面极小，适合静态 Go 二进制
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/server /server

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
