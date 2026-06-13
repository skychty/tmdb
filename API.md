# API 使用文档

本文档描述 TMDB 影片信息服务器对外提供的 REST API，包括请求参数、响应头、返回字段及示例。

**Base URL 示例：**

```
https://tmdb.blogsite.org
```

**API 版本前缀：** `/api/v1`

---

## 通用说明

### 请求方式

所有业务接口均为 **GET**，响应格式为 **JSON**（`Content-Type: application/json`）。

### 通用 Query 参数

适用于所有 `/api/v1/movies/*` 与 `/api/v1/tv/*` 接口：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `region` | string | 否 | 自动识别 | ISO 3166-1 两位国家/地区代码，如 `CN`、`US`、`JP`、`GB`（英国请用 `GB`，不要用 `UK`） |
| `language` | string | 否 | `zh-CN` | 语言代码，控制标题、简介等翻译，如 `zh-CN`、`en-US`、`ja-JP` |
| `page` | int | 否 | `1` | 页码，从 1 开始 |

**region 自动识别规则：**

- 未传 `region` 时，服务器根据客户端公网 IP 解析国家代码
- 内网 IP 或解析失败时，使用服务端配置的 `DEFAULT_REGION`（默认 `CN`）

### 响应头

| 响应头 | 说明 | 示例 |
|--------|------|------|
| `X-Region` | 实际使用的区域代码 | `CN` |
| `X-Region-Source` | 区域来源 | `query`（请求参数指定）或 `ip`（IP 自动识别） |
| `Access-Control-Allow-Origin` | CORS | `*` |

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| `200` | 成功 |
| `400` | 请求参数错误（如 region 格式非法、page 非正整数） |
| `502` | TMDB 请求失败，且无可用过期缓存 |

### 错误响应格式

```json
{
  "error": "region must be a 2-letter ISO 3166-1 code"
}
```

### 缓存与降级

- 新鲜缓存默认 **24 小时**（`CACHE_TTL`）
- 过期缓存保留 **7 天**（`STALE_CACHE_TTL`），在 TMDB 限流排队超时或上游失败时用于降级返回
- 降级成功时仍返回 **200**，数据可能来自过期缓存

---

## 系统接口

### GET /health

健康检查。

**响应示例：**

```json
{
  "status": "ok"
}
```

### GET /

浏览器首页（HTML），包含 API 快捷链接，非 JSON 接口。

---

## 电影接口

### GET /api/v1/movies/latest

获取指定地区**最新上线**电影（数据源：TMDB `/movie/now_playing`）。

**适用场景：** 当地影院正在上映或近期上映的影片。

**请求示例：**

```
GET /api/v1/movies/latest?region=CN&language=zh-CN&page=1
```

---

### GET /api/v1/movies/popular

获取**全球热门**电影（数据源：TMDB `/movie/popular`）。

**说明：** 该接口按全球热度排序，不同 `region` 返回列表**相近**，主要差异在 `release_date` 等字段。

**请求示例：**

```
GET /api/v1/movies/popular?region=CN&language=zh-CN&page=1
```

---

### GET /api/v1/movies/regional-popular

获取**地区热门**电影（数据源：TMDB `/discover/movie`）。

**说明：** 筛选指定地区近 3 个月院线上映（`with_release_type=2|3`）的影片，按热度排序。不同 `region` 返回列表**差异明显**。

**请求示例：**

```
GET /api/v1/movies/regional-popular?region=CN&language=zh-CN&page=1
```

---

## 连续剧接口

### GET /api/v1/tv/on-the-air

获取**正在播出**的连续剧（数据源：TMDB `/tv/on_the_air`）。

**说明：** 近 7 天内有集数更新的剧集。

**请求示例：**

```
GET /api/v1/tv/on-the-air?region=CN&language=zh-CN&page=1
```

---

### GET /api/v1/tv/popular

获取**全球热门**连续剧（数据源：TMDB `/tv/popular`）。

**请求示例：**

```
GET /api/v1/tv/popular?region=CN&language=zh-CN&page=1
```

---

### GET /api/v1/tv/regional-popular

获取**地区热门**连续剧（数据源：TMDB `/discover/tv`）。

**说明：** 按 `with_origin_country` 筛选该地区制作的剧集，近 3 个月首播，按热度排序。不同国家/地区列表**差异明显**。

**请求示例：**

```
GET /api/v1/tv/regional-popular?region=JP&language=ja-JP&page=1
```

---

## 列表响应结构（电影 / 连续剧通用）

电影与连续剧列表接口返回结构类似，均包含分页信息与 `results` 数组。

### 电影列表响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `page` | int | 当前页码 |
| `total_pages` | int | 总页数 |
| `total_results` | int | 总结果数 |
| `region` | string | 实际使用的区域代码（大写） |
| `cached_at` | string | 本次数据写入缓存的时间（UTC，RFC3339 格式） |
| `results` | array | 电影列表，见下表 |

### 连续剧列表响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `page` | int | 当前页码 |
| `total_pages` | int | 总页数 |
| `total_results` | int | 总结果数 |
| `region` | string | 实际使用的区域代码（大写） |
| `cached_at` | string | 本次数据写入缓存的时间（UTC，RFC3339 格式） |
| `results` | array | 连续剧列表，见下表 |

---

## 电影对象字段（results[]）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int | TMDB 电影 ID，可用于拼接详情页或跳转 |
| `title` | string | 本地化标题（受 `language` 参数影响） |
| `original_title` | string | 原始语言标题 |
| `overview` | string | 简介/剧情概要 |
| `release_date` | string | 上映日期，格式 `YYYY-MM-DD`；无数据时为空字符串 |
| `poster_url` | string | **竖版海报**完整 URL（宽 500px）；无海报时为空字符串 |
| `backdrop_url` | string | **横版背景图**完整 URL（原始尺寸）；无图时为空字符串 |
| `vote_average` | float | 平均评分（0–10） |
| `vote_count` | int | 评分人数 |
| `popularity` | float | TMDB 热度值（越大越热门） |
| `genre_ids` | int[] | 类型 ID 数组，见 [类型 ID 对照](#类型-id-对照) |

**图片说明：**

- `poster_url`：竖向海报（约 2:3）
- `backdrop_url`：横向宽图（约 16:9），适合横幅/背景展示
- TMDB 列表接口不提供单独的「横版海报」字段

---

## 连续剧对象字段（results[]）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int | TMDB 连续剧 ID |
| `name` | string | 本地化剧名（受 `language` 参数影响） |
| `original_name` | string | 原始语言剧名 |
| `overview` | string | 简介/剧情概要 |
| `first_air_date` | string | 首播日期，格式 `YYYY-MM-DD`；无数据时为空字符串 |
| `poster_url` | string | **竖版海报**完整 URL（宽 500px） |
| `backdrop_url` | string | **横版背景图**完整 URL（原始尺寸） |
| `vote_average` | float | 平均评分（0–10） |
| `vote_count` | int | 评分人数 |
| `popularity` | float | TMDB 热度值 |
| `genre_ids` | int[] | 类型 ID 数组 |
| `origin_country` | string[] | 制作国家/地区代码数组，如 `["US"]`、`["KR"]`、`["CN"]` |

---

## 响应示例

### 电影列表示例

```json
{
  "page": 1,
  "total_pages": 3,
  "total_results": 42,
  "region": "CN",
  "cached_at": "2026-06-12T10:00:00Z",
  "results": [
    {
      "id": 1228710,
      "title": "星球大战：曼达洛人与古古",
      "original_title": "The Mandalorian and Grogu",
      "overview": "“曼达洛人”丁·贾伦和“尤达宝宝”古古的星际冒险全面升级……",
      "release_date": "2026-05-22",
      "poster_url": "https://image.tmdb.org/t/p/w500/cafHjxEvhslX9MpBMxmPWxC5GWB.jpg",
      "backdrop_url": "https://image.tmdb.org/t/p/original/6zg7A9ICOthNR2TSXlT51KvXrsA.jpg",
      "vote_average": 6.8,
      "vote_count": 437,
      "popularity": 349.3,
      "genre_ids": [28, 12, 878]
    }
  ]
}
```

### 连续剧列表示例

```json
{
  "page": 1,
  "total_pages": 15,
  "total_results": 291,
  "region": "JP",
  "cached_at": "2026-06-12T10:00:00Z",
  "results": [
    {
      "id": 290019,
      "name": "和班上第二可爱的女孩子成为了朋友",
      "original_name": "クラスで2番目に可愛い女の子と友だちになった",
      "overview": "……",
      "first_air_date": "2026-04-01",
      "poster_url": "https://image.tmdb.org/t/p/w500/xxx.jpg",
      "backdrop_url": "https://image.tmdb.org/t/p/original/xxx.jpg",
      "vote_average": 8.5,
      "vote_count": 120,
      "popularity": 98.2,
      "genre_ids": [16, 35],
      "origin_country": ["JP"]
    }
  ]
}
```

---

## 接口对比速查

| 接口 | 内容类型 | 数据源 | region 差异 |
|------|----------|--------|-------------|
| `/movies/latest` | 电影 | now_playing | 有（上映日） |
| `/movies/popular` | 电影 | popular | 小（全球热度榜） |
| `/movies/regional-popular` | 电影 | discover/movie | **大**（当地院线上映） |
| `/tv/on-the-air` | 连续剧 | tv/on_the_air | 较小 |
| `/tv/popular` | 连续剧 | tv/popular | 较小 |
| `/tv/regional-popular` | 连续剧 | discover/tv | **大**（当地制作） |

---

## 类型 ID 对照

`genre_ids` 为 TMDB 标准类型 ID，常见值如下：

| ID | 英文 | 中文 |
|----|------|------|
| 28 | Action | 动作 |
| 12 | Adventure | 冒险 |
| 16 | Animation | 动画 |
| 35 | Comedy | 喜剧 |
| 80 | Crime | 犯罪 |
| 99 | Documentary | 纪录片 |
| 18 | Drama | 剧情 |
| 10751 | Family | 家庭 |
| 14 | Fantasy | 奇幻 |
| 36 | History | 历史 |
| 27 | Horror | 恐怖 |
| 10402 | Music | 音乐 |
| 9648 | Mystery | 悬疑 |
| 10749 | Romance | 爱情 |
| 878 | Science Fiction | 科幻 |
| 53 | Thriller | 惊悚 |
| 10752 | War | 战争 |
| 37 | Western | 西部 |

完整列表见 [TMDB Genre API](https://developer.themoviedb.org/reference/genre-movie-list)。

---

## 常用 region 代码

| 代码 | 地区 |
|------|------|
| `CN` | 中国大陆 |
| `HK` | 香港 |
| `TW` | 台湾 |
| `JP` | 日本 |
| `KR` | 韩国 |
| `IN` | 印度 |
| `US` | 美国 |
| `GB` | 英国 |

---

## 设备端调用示例

### cURL

```bash
# 自动识别 region
curl "https://tmdb.blogsite.org/api/v1/movies/latest"

# 指定 region 和 language
curl "https://tmdb.blogsite.org/api/v1/movies/regional-popular?region=CN&language=zh-CN&page=1"

# 连续剧
curl "https://tmdb.blogsite.org/api/v1/tv/regional-popular?region=JP&language=ja-JP"
```

### 分页

```bash
curl "https://tmdb.blogsite.org/api/v1/movies/popular?region=CN&page=2"
```

当 `page` > `total_pages` 时，TMDB 返回空列表或错误，建议客户端根据 `total_pages` 限制翻页。

---

## 注意事项

1. **无需 API Key**：设备端调用本服务器接口不需要 TMDB Token，Token 由服务端持有。
2. **图片 URL 可直接使用**：`poster_url`、`backdrop_url` 已是完整 HTTPS 地址。
3. **空字段**：无海报/背景时对应 URL 为空字符串 `""`，客户端应做占位图处理。
4. **日期字段**：无日期时为 `""`，不是 `null`。
5. **评分**：`vote_average` 为 0 且 `vote_count` 为 0 表示尚无评分。
6. **高并发**：服务端有缓存与限流，设备端建议本地缓存列表数据，避免频繁重复请求同一 URL。
