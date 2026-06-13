package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	host := c.Request.Host
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>TMDB 影片信息服务器</title>
  <style>
    body { font-family: sans-serif; max-width: 720px; margin: 40px auto; padding: 0 16px; line-height: 1.6; }
    a { color: #2563eb; }
    code { background: #f3f4f6; padding: 2px 6px; border-radius: 4px; }
  </style>
</head>
<body>
  <h1>TMDB 影片信息服务器</h1>
  <p>服务运行中。局域网内其它设备可通过 <code>http://%s</code> 访问。</p>
  <h2>电影 API</h2>
  <ul>
    <li><a href="/api/v1/movies/latest">最新上线（自动识别区域）</a></li>
    <li><a href="/api/v1/movies/popular">全球热门（TMDB popular）</a></li>
    <li><a href="/api/v1/movies/regional-popular?region=CN">地区热门（TMDB discover）</a></li>
  </ul>
  <h2>连续剧 API</h2>
  <ul>
    <li><a href="/api/v1/tv/on-the-air?region=CN">正在播出</a></li>
    <li><a href="/api/v1/tv/popular?region=CN">全球热门剧集</a></li>
    <li><a href="/api/v1/tv/regional-popular?region=CN">地区热门剧集</a></li>
  </ul>
  <h2>其它</h2>
  <ul>
    <li><a href="/health">健康检查</a></li>
  </ul>
  <h2>参数说明</h2>
  <ul>
    <li><code>region</code>：可选，国家/地区代码，如 CN、US；未指定时根据客户端 IP 自动识别</li>
    <li><code>language</code>：可选，默认 en-US</li>
    <li><code>page</code>：可选，默认 1</li>
  </ul>
</body>
</html>`, host)))
}
