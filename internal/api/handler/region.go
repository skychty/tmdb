package handler

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"tmdb/internal/geoip"
)

var regionPattern = regexp.MustCompile(`^[A-Za-z]{2}$`)

func resolveRegion(c *gin.Context, geoIP *geoip.Resolver) (region, source string, err error) {
	queryRegion := strings.TrimSpace(c.Query("region"))
	if queryRegion != "" {
		if !regionPattern.MatchString(queryRegion) {
			return "", "", errInvalidRegion
		}
		return strings.ToUpper(queryRegion), "query", nil
	}

	if cfRegion := geoip.NormalizeCountryCode(c.GetHeader("CF-IPCountry")); cfRegion != "" {
		return cfRegion, "cloudflare", nil
	}

	ip := clientIP(c)
	if geoip.IsLocalIP(ip) {
		return geoIP.DefaultRegion(), "default", nil
	}

	if region, ok := geoIP.LookupCountry(ip); ok {
		return region, "geolite2", nil
	}

	return geoIP.DefaultRegion(), "default", nil
}

func clientIP(c *gin.Context) string {
	if ip := strings.TrimSpace(c.GetHeader("CF-Connecting-IP")); ip != "" {
		return ip
	}
	if ip := strings.TrimSpace(c.GetHeader("X-Real-IP")); ip != "" {
		return ip
	}
	return c.ClientIP()
}
