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
	if geoip.IsLocalIP(ip) {
		result.region = geoIP.DefaultRegion()
		result.source = "default"
		logRegionResolve(c, result)
		return result, nil
	}

	if region, ok := geoIP.LookupCountry(ip); ok {
		result.region = region
		result.source = "geolite2"
		logRegionResolve(c, result)
		return result, nil
	}

	result.region = geoIP.DefaultRegion()
	result.source = "default"
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
		if ip := strings.TrimSpace(c.GetHeader("CF-Connecting-IP")); ip != "" {
			return ip
		}
	}
	if ip := strings.TrimSpace(c.GetHeader("X-Real-IP")); ip != "" {
		return ip
	}
	if ip := remoteHost(c.Request.RemoteAddr); ip != "" && !geoip.IsLocalIP(ip) {
		return ip
	}
	return c.ClientIP()
}

func remoteHost(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return strings.TrimSpace(remoteAddr)
	}
	return strings.TrimSpace(host)
}
