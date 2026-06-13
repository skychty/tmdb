# TMDB 影片信息服务器

从 TMDB 拉取按国家/地区过滤的影片数据，并通过 Redis 缓存 24 小时，向设备端提供 REST API。

## 功能

- `GET /` — 浏览器首页（含 API 快捷链接）
- `GET /api/v1/movies/latest` — 本地最新上线影片（对应 TMDB `now_playing`）
- `GET /api/v1/movies/popular` — 全球热门影片（TMDB `popular`，各 region 列表相近）
- `GET /api/v1/movies/regional-popular` — 地区热门影片（TMDB `discover`，按 region 筛选当地院线上映）
- `GET /api/v1/tv/on-the-air` — 正在播出的连续剧（TMDB `tv/on_the_air`）
- `GET /api/v1/tv/popular` — 全球热门连续剧（TMDB `tv/popular`）
- `GET /api/v1/tv/regional-popular` — 地区热门连续剧（TMDB `discover/tv`，按 `with_origin_country` 筛选当地制作）
- `GET /health` — 健康检查

## 公网部署

完整公网部署步骤见 **[DEPLOY.md](DEPLOY.md)**（含 Nginx、HTTPS、安全组、运维命令）。

## 快速启动

1. 复制环境变量文件并填入 TMDB Token：

```bash
cp .env.example .env
# 编辑 .env，设置 TMDB_ACCESS_TOKEN
```

2. 使用 Docker Compose 启动：

```bash
docker compose up --build
```

3. 测试 API：

```bash
curl "http://localhost:8080/api/v1/movies/latest?region=CN&language=zh-CN"
curl "http://localhost:8080/api/v1/movies/popular?region=CN&language=zh-CN"
curl "http://localhost:8080/api/v1/movies/regional-popular?region=CN&language=zh-CN"
curl "http://localhost:8080/api/v1/tv/on-the-air?region=CN&language=zh-CN"
curl "http://localhost:8080/api/v1/tv/popular?region=CN&language=zh-CN"
curl "http://localhost:8080/api/v1/tv/regional-popular?region=CN&language=zh-CN"
```

## 局域网访问

服务默认监听 `0.0.0.0:8080`，局域网内其它设备可通过服务器 IP 访问：

```bash
# 查看本机 IP（示例输出 10.0.0.11）
hostname -I

# 在其它设备的浏览器中打开
http://10.0.0.11:8080/
http://10.0.0.11:8080/api/v1/movies/latest?region=CN&language=zh-CN
```

若无法访问，请检查防火墙是否放行 8080 端口：

```bash
sudo ufw allow 8080/tcp
```

## 本地开发（无 Docker）

需要 Go 1.21+ 和 Redis。

```bash
go mod tidy
go run ./cmd/server
```

## 请求参数

| 参数 | 必填 | 说明 |
|------|------|------|
| region | 否 | ISO 3166-1 两位国家/地区代码，如 `CN`；未指定时根据客户端 IP 自动识别 |
| language | 否 | 语言，默认 `zh-CN` |
| page | 否 | 页码，默认 `1` |

响应头会返回区域来源：
- `X-Region`：实际使用的区域代码
- `X-Region-Source`：`query`（API 指定）或 `ip`（IP 自动识别）

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| TMDB_ACCESS_TOKEN | TMDB Bearer Token | - |
| TMDB_API_KEY | TMDB API Key（可选，二选一） | - |
| TMDB_BASE_URL | TMDB API 地址 | `https://api.themoviedb.org/3` |
| TMDB_IMAGE_BASE | 图片 CDN 前缀 | `https://image.tmdb.org/t/p` |
| REDIS_ADDR | Redis 地址 | `127.0.0.1:6379` |
| CACHE_TTL | 新鲜缓存有效期 | `24h` |
| STALE_CACHE_TTL | 过期缓存保留时间（用于降级） | `168h` |
| TMDB_RATE_LIMIT | TMDB 全局限流（请求/秒） | `40` |
| TMDB_RATE_BURST | TMDB 限流突发容量 | `40` |
| TMDB_QUEUE_TIMEOUT | 排队等待 TMDB 令牌超时 | `5s` |
| GEOIP_CACHE_TTL | IP 区域识别缓存有效期 | `24h` |
| DEFAULT_REGION | IP 无法识别时的默认区域（含内网 IP） | `CN` |
| HTTP_HOST | 监听地址（`0.0.0.0` 允许局域网访问） | `0.0.0.0` |
| HTTP_PORT | 服务端口 | `8080` |

## 缓存与限流策略

- 电影 Key：`tmdb:movies:{latest|popular|regional-popular}:{region}:{language}:{page}`
- 连续剧 Key：`tmdb:tv:{on-the-air|popular|regional-popular-v2}:{region}:{language}:{page}`
- **新鲜缓存 TTL**：24 小时（`CACHE_TTL`）
- **过期缓存保留**：7 天（`STALE_CACHE_TTL`），用于降级
- **TMDB 全局限流**：令牌桶，默认 40 req/s（`TMDB_RATE_LIMIT`）
- **排队超时**：默认 5 秒（`TMDB_QUEUE_TIMEOUT`），超时后返回过期缓存；无过期缓存才 502
- 同一 key 并发 miss 时使用 singleflight 合并 TMDB 请求
