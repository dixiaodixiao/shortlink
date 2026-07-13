# 短链接服务 —— 设计文档 (DESIGN.md)

> 一个生产级思路的 URL 短链接服务，用于完整体验后端研发全流程。
> 领域简单，但在缓存、并发、发号、可观测性上层层加深，兼作系统设计复习题。

## 1. 目标与范围

**v1（先跑通）**
- 提交长链接，生成短码
- 通过短码 302 重定向到原始链接
- Redis 缓存热点短码，扛读流量
- 点击计数

**v2（后续增量）**
- 接口限流
- 自定义别名
- 短链过期时间
- 基础指标 / 可观测性

## 2. 技术选型

| 关注点 | 选型 | 理由 |
|--------|------|------|
| 语言 | Go 1.26 | 后端主流，并发模型好讲 |
| Web 框架 | Gin | 国内岗位提及率高，中间件生态成熟 |
| 持久化 | PostgreSQL | 存储 code ↔ 原始 URL 映射与元数据 |
| 缓存 | Redis | 短链读远多于写，缓存热点短码 |
| 驱动 | pgx / go-redis | 社区主流 |

## 3. 短码生成方案（核心考点）

**方案：数据库自增 ID → Base62 编码**

- 写入映射时拿到自增主键 ID（如 `10000`）
- 用 Base62（`0-9a-zA-Z`，共 62 字符）编码为短字符串（如 `2Bi`）
- 短、无碰撞、实现简单

**可演进性（面试加分点）**

发号策略抽象为接口 `IDGenerator`，v1 用数据库自增实现；未来可无缝替换为：
- 雪花算法（Snowflake）——分布式、趋势递增
- 号段模式（Segment）——批量取号，减少 DB 压力

## 4. 分层架构

```
HTTP 请求
   │
   ▼
handler   （HTTP 层：参数解析、响应、状态码）
   │
   ▼
service   （业务逻辑：base62 编码、发号、缓存读写策略）
   │
   ▼
repository（数据访问：Postgres + Redis）
   │
   ▼
PostgreSQL / Redis
```

- 依赖方向单向向下，便于单元测试（各层可 mock 下层）
- service 层不感知 HTTP，也不直接拼 SQL

## 5. 目录结构（Go 社区惯例）

```
cmd/server/main.go        程序入口，装配依赖
internal/handler/         HTTP 处理器（Gin）
internal/service/         业务逻辑（base62、发号、缓存策略）
internal/repository/      Postgres + Redis 访问
internal/model/           领域数据结构
config/                   配置加载
migrations/               建表 SQL
DESIGN.md                 本文档
```

## 6. 数据模型

`links` 表：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PK | 自增主键，Base62 编码来源 |
| code | VARCHAR UNIQUE | 短码（编码后的 id） |
| original_url | TEXT | 原始长链接 |
| click_count | BIGINT | 点击次数 |
| created_at | TIMESTAMPTZ | 创建时间 |

> 注：`code` 由 `id` 编码得到，理论上可不落库、按需计算；此处冗余存储以便按 code 建唯一索引、简化查询。

## 7. API 契约（v1）

| 方法 | 路径 | 请求 | 响应 |
|------|------|------|------|
| POST | `/api/links` | `{"url": "https://..."}` | `201 {"code":"2Bi","short_url":"http://host/2Bi"}` |
| GET | `/{code}` | — | `302` 重定向到原始链接 |
| GET | `/api/links/{code}` | — | `200` 短链详情（含 click_count） |
| GET | `/healthz` | — | `200 {"status":"ok"}` |

**错误约定**：`400` 参数非法，`404` 短码不存在，`500` 服务端错误；响应体统一 `{"error":"..."}`。

## 8. 缓存策略

- 读（重定向）：先查 Redis `code -> url`；命中直接重定向；未命中查 DB、回填 Redis（设 TTL）
- 写（创建）：写 DB 后按需回填缓存
- 点击计数：先在 Redis 累加，异步/定期落库，避免每次重定向都写 DB（v2 完善）

## 9. 里程碑

1. 骨架跑通空服务（`/healthz` 可访问）
2. Base62 编码 + 单元测试
3. 创建短链（DB 落库）
4. 重定向（DB 查询）
5. Redis 缓存接入
6. 点击计数
7. 质量把关（code-review / security-review）
8. Docker + CI/CD
9. 文档 + 推送 GitHub
