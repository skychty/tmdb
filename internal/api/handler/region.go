package handler

import (
	"log"
	"net"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"tmdb/internal/geoip"
)

var regionPattern = regexp.MustCompile(`^[A-Za-z]{2}$`)

type regionResolveResult struct {
	region   string
	source   string
	clientIP string
}

func resolveRegion(c *gin.Context, geoIP *geoip.Resolver) (regionResolveResult, error) {
	queryRegion := strings.TrimSpace(c.Query("region"))
	if queryRegion != "" {
		if !regionPattern.MatchString(queryRegion) {
			return regionResolveResult{}, errInvalidRegion
		}
		return regionResolveResult{
			region: strings.ToUpper(queryRegion),
			source: "query",
		}, nil
	}

	if isCloudflareRequest(c) {
		if cfRegion := geoip.NormalizeCountryCode(c.GetHeader("CF-IPCountry")); cfRegion != "" {
			result := regionResolveResult{
				region:   cfRegion,
				source:   "cloudflare",
				clientIP: strings.TrimSpace(c.GetHeader("CF-Connecting-IP")),
			}
			logRegionResolve(c, result)
			return result, nil
		}
	}

	ip := clientIP(c)
	result := regionResolveResult{clientIP: ip}

	lookup := geoIP.ResolveRegion(c.Request.Context(), ip)
	result.region = lookup.Region
	result.source = lookup.Source
	logRegionResolve(c, result)
	return result, nil
}

func applyRegionHeaders(c *gin.Context, result regionResolveResult) {
	c.Header("X-Region", result.region)
	c.Header("X-Region-Source", result.source)
	if result.clientIP != "" {
		c.Header("X-Client-IP", result.clientIP)
	}
}

func logRegionResolve(c *gin.Context, result regionResolveResult) {
	log.Printf(
		"region auto-detect: region=%s source=%s client_ip=%s cf_ipcountry=%q cf_ray=%q x_real_ip=%q x_forwarded_for=%q remote_addr=%s",
		result.region,
		result.source,
		result.clientIP,
		c.GetHeader("CF-IPCountry"),
		c.GetHeader("CF-Ray"),
		c.GetHeader("X-Real-IP"),
		c.GetHeader("X-Forwarded-For"),
		c.Request.RemoteAddr,
	)
}

func isCloudflareRequest(c *gin.Context) bool {
	return strings.TrimSpace(c.GetHeader("CF-Ray")) != ""
}

func clientIP(c *gin.Context) string {
	if isCloudflareRequest(c) {
		if ip := strings.TrimSpace(c.GetHeader("CF-Connecting-IP")); isPublicIP(ip) {
			return ip
		}
	}

	if ip := pickClientIP(c.GetHeader("X-Real-IP"), c.GetHeader("X-Forwarded-For")); ip != "" {
		return ip
	}

	if ip := remoteHost(c.Request.RemoteAddr); isPublicIP(ip) {
		return ip
	}

	return c.ClientIP()
}

// pickClientIP resolves the original client IP from proxy headers.
// When multiple proxies are involved, X-Real-IP is usually the immediate upstream
// (last X-Forwarded-For hop) and the leftmost X-Forwarded-For entry is the client.
func pickClientIP(xRealIP, xForwardedFor string) string {
	xRealIP = strings.TrimSpace(xRealIP)
	xffIPs := parseForwardedFor(xForwardedFor)

	if len(xffIPs) >= 2 && xRealIP != "" && xffIPs[len(xffIPs)-1] == xRealIP {
		if isPublicIP(xffIPs[0]) {
			return xffIPs[0]
		}
	}

	if len(xffIPs) == 1 && isPublicIP(xffIPs[0]) {
		return xffIPs[0]
	}

	if isPublicIP(xRealIP) {
		return xRealIP
	}

	return ""
}

func parseForwardedFor(xForwardedFor string) []string {
	parts := strings.Split(xForwardedFor, ",")
	ips := make([]string, 0, len(parts))
	for _, part := range parts {
		ip := strings.TrimSpace(part)
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

func isPublicIP(ip string) bool {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return false
	}
	return !geoip.IsLocalIP(ip)
}

func remoteHost(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return strings.TrimSpace(remoteAddr)
	}
	return strings.TrimSpace(host)
}
