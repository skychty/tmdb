package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"tmdb/internal/geoip"
)

func resolveRegion(c *gin.Context, geoIP *geoip.Resolver) (region, source string, err error) {
	queryRegion := strings.TrimSpace(c.Query("region"))
	if queryRegion != "" {
		if !regionPattern.MatchString(queryRegion) {
			return "", "", errInvalidRegion
		}
		return strings.ToUpper(queryRegion), "query", nil
	}

	region, err = geoIP.ResolveRegion(c.Request.Context(), c.ClientIP())
	if err != nil {
		return "", "", err
	}
	return region, "ip", nil
}
