# 公网部署手册

本文档说明如何将 [TMDB 影片信息服务器](https://github.com/skychty/tmdb) 部署到公网，供设备端通过 HTTPS 访问。

## 架构概览

```
设备/浏览器
    │  HTTPS :443
    ▼
  Nginx（反向代理 + SSL）
    │  http://127.0.0.1:8080
    ▼
  Go API（Docker）
    ├── Redis（Docker 内网，不暴露公网）
    ├── TMDB API
    └── ip-api.com（IP 区域识别）
```

---

## 一、服务器要求

| 项目 | 建议 |
|------|------|
| 系统 | Ubuntu 22.04 / Debian 12 |
| 配置 | 1 核 CPU、1GB 内存（最低） |
| 软件 | Docker、Docker Compose、Nginx、Certbot |
| 域名 | 可选但强烈推荐（如 `api.example.com`） |
| 开放端口 | 22（SSH）、80（HTTP）、443（HTTPS） |

**不要**对公网开放 8080、6379 端口。

---

## 二、云厂商安全组

在云控制台（阿里云 / 腾讯云 / AWS 等）配置入站规则：

| 端口 | 协议 | 来源 | 说明 |
|------|------|------|------|
| 22 | TCP | 你的 IP | SSH 管理 |
| 80 | TCP | 0.0.0.0/0 | HTTP（证书申请 + 跳转 HTTPS） |
| 443 | TCP | 0.0.0.0/0 | HTTPS 对外服务 |

---

## 三、安装基础环境

以 Ubuntu 为例：

```bash
sudo apt update
sudo apt install -y docker.io docker-compose-v2 git nginx certbot python3-certbot-nginx
sudo systemctl enable --now docker nginx
sudo usermod -aG docker $USER
```

重新登录 SSH 使 `docker` 组生效。

---

## 四、拉取代码

```bash
sudo mkdir -p /opt/tmdb
sudo chown $USER:$USER /opt/tmdb
cd /opt/tmdb

# SSH 方式（推荐）
git clone git@github.com:skychty/tmdb.git .

# 或 HTTPS 方式
# git clone https://github.com/skychty/tmdb.git .
```

---

## 五、配置环境变量

```bash
cd /opt/tmdb
cp .env.example .env
nano .env
```

`.env` 示例：

```bash
TMDB_ACCESS_TOKEN=你的BearerToken
TMDB_BASE_URL=https://api.themoviedb.org/3
TMDB_IMAGE_BASE=https://image.tmdb.org/t/p
REDIS_ADDR=redis:6379
CACHE_TTL=24h
GEOIP_CACHE_TTL=24h
DEFAULT_REGION=CN
HTTP_HOST=0.0.0.0
HTTP_PORT=8080
```

> **注意：** `.env` 含密钥，不要提交到 Git。Token 在 [TMDB API 设置页](https://www.themoviedb.org/settings/api) 获取。

---

## 六、启动服务（生产模式）

项目提供 `docker-compose.prod.yml`，特点：
- Redis **不**映射到公网
- API 仅绑定 `127.0.0.1:8080`，由 Nginx 对外暴露

```bash
cd /opt/tmdb
docker compose -f docker-compose.prod.yml up -d --build
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs -f app
```

验证本机服务：

```bash
curl http://127.0.0.1:8080/health
# 期望：{"status":"ok"}
```

---

## 七、配置域名 DNS

在域名服务商添加 A 记录：

| 类型 | 主机记录 | 记录值 |
|------|----------|--------|
| A | `api`（或 `@`） | 服务器公网 IP |

示例：`api.example.com` → `1.2.3.4`

等待 DNS 生效（通常几分钟）：

```bash
dig +short api.example.com
```

---

## 八、Nginx 反向代理

复制项目自带配置模板：

```bash
sudo cp /opt/tmdb/deploy/nginx-tmdb.conf /etc/nginx/sites-available/tmdb
sudo nano /etc/nginx/sites-available/tmdb
```

将 `api.example.com` 改为你的域名，然后启用：

```bash
sudo ln -sf /etc/nginx/sites-available/tmdb /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl reload nginx
```

此时可通过 HTTP 访问：`http://api.example.com/health`

---

## 九、申请 HTTPS 证书

```bash
sudo certbot --nginx -d api.example.com
```

按提示选择自动重定向 HTTP → HTTPS。证书会自动续期。

验证：

```bash
curl https://api.example.com/health
```

---

## 十、系统防火墙

```bash
sudo ufw allow OpenSSH
sudo ufw allow 'Nginx Full'
sudo ufw enable
sudo ufw status
```

---

## 十一、公网 API 验证

```bash
# 健康检查
curl https://api.example.com/health

# 自动识别区域（公网 IP）
curl -s -D - "https://api.example.com/api/v1/movies/latest" -o /dev/null | grep X-Region

# 指定区域
curl "https://api.example.com/api/v1/movies/latest?region=CN&language=zh-CN"
curl "https://api.example.com/api/v1/movies/popular?region=US&language=en-US"
```

浏览器访问：`https://api.example.com/`

### 响应头说明

| 响应头 | 含义 |
|--------|------|
| `X-Region` | 实际使用的区域代码（如 `CN`） |
| `X-Region-Source` | `query`（API 指定）或 `ip`（IP 自动识别） |

---

## 十二、设备端接入

设备端请求示例：

```
GET https://api.example.com/api/v1/movies/latest
GET https://api.example.com/api/v1/movies/popular?region=CN&language=zh-CN&page=1
```

| 参数 | 必填 | 说明 |
|------|------|------|
| region | 否 | 国家/地区代码；未指定时根据客户端公网 IP 自动识别 |
| language | 否 | 默认 `zh-CN` |
| page | 否 | 默认 `1` |

---

## 十三、日常运维

```bash
cd /opt/tmdb

# 查看状态
docker compose -f docker-compose.prod.yml ps

# 查看日志
docker compose -f docker-compose.prod.yml logs -f app

# 更新部署
git pull
docker compose -f docker-compose.prod.yml up -d --build

# 重启应用
docker compose -f docker-compose.prod.yml restart app

# 停止服务
docker compose -f docker-compose.prod.yml down
```

### 开机自启

Docker 服务默认开机自启。Compose 中已配置 `restart: unless-stopped`，服务器重启后容器会自动恢复。

---

## 十四、故障排查

| 现象 | 排查步骤 |
|------|----------|
| 外网无法访问 | 检查云安全组是否放行 80/443；`sudo ufw status` |
| 502 Bad Gateway | `docker compose -f docker-compose.prod.yml ps` 确认 app 在运行；`curl http://127.0.0.1:8080/health` |
| TMDB 请求失败 | 检查 `.env` 中 `TMDB_ACCESS_TOKEN` 是否有效 |
| 区域识别不准 | 确认 Nginx 传递了 `X-Real-IP`；公网 IP 才会走 GeoIP |
| Redis 连接失败 | 确认 `REDIS_ADDR=redis:6379`（Compose 内网地址） |

---

## 十五、安全清单

- [ ] `.env` 未提交到 Git
- [ ] Redis 6379 未暴露公网
- [ ] API 8080 仅绑定 127.0.0.1
- [ ] 已启用 HTTPS
- [ ] 云安全组仅开放必要端口
- [ ] TMDB Token 定期轮换（泄露时在 TMDB 后台重新生成）

---

## 十六、快速命令备忘

```bash
# 一键部署（首次）
cd /opt/tmdb && cp .env.example .env && nano .env
docker compose -f docker-compose.prod.yml up -d --build

# 一键更新
cd /opt/tmdb && git pull && docker compose -f docker-compose.prod.yml up -d --build
```
