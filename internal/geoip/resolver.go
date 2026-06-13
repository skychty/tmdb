package geoip

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
	"golang.org/x/sync/singleflight"

	"tmdb/internal/cache"
)

type LookupResult struct {
	Region string
	Source string
}

type Resolver struct {
	db            *geoip2.Reader
	cache         cache.Cache
	defaultRegion string
	cacheTTL      time.Duration
	httpClient    *http.Client
	group         singleflight.Group
}

func NewResolver(dbPath, defaultRegion string, redisCache cache.Cache, cacheTTL time.Duration) (*Resolver, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open geoip db %s: %w", dbPath, err)
	}
	return &Resolver{
		db:            db,
		cache:         redisCache,
		defaultRegion: strings.ToUpper(defaultRegion),
		cacheTTL:      cacheTTL,
		httpClient:    newIPAPIClient(),
	}, nil
}

func (r *Resolver) Close() error {
	return r.db.Close()
}

func (r *Resolver) DefaultRegion() string {
	return r.defaultRegion
}

func IsLocalIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return true
	}
	return parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsUnspecified()
}

func (r *Resolver) ResolveRegion(ctx context.Context, ip string) LookupResult {
	if IsLocalIP(ip) {
		return LookupResult{Region: r.defaultRegion, Source: "default"}
	}

	cacheKey := "geoip:region:" + ip
	if data, ok, err := r.cache.Get(ctx, cacheKey); err == nil && ok {
		region := strings.ToUpper(strings.TrimSpace(string(data)))
		if IsValidCountryCode(region) {
			return LookupResult{Region: region, Source: "cache"}
		}
	}

	result, _, _ := r.group.Do(ip, func() (any, error) {
		if data, ok, err := r.cache.Get(ctx, cacheKey); err == nil && ok {
			region := strings.ToUpper(strings.TrimSpace(string(data)))
			if IsValidCountryCode(region) {
				return LookupResult{Region: region, Source: "cache"}, nil
			}
		}

		if region, err := lookupIPAPI(ctx, ip, r.httpClient); err == nil {
			_ = r.cache.Set(ctx, cacheKey, []byte(region), r.cacheTTL)
			return LookupResult{Region: region, Source: "ip-api"}, nil
		}

		if region, ok := r.lookupGeoLite2(ip); ok {
			_ = r.cache.Set(ctx, cacheKey, []byte(region), r.cacheTTL)
			return LookupResult{Region: region, Source: "geolite2"}, nil
		}

		return LookupResult{Region: r.defaultRegion, Source: "default"}, nil
	})

	lookup, ok := result.(LookupResult)
	if !ok {
		return LookupResult{Region: r.defaultRegion, Source: "default"}
	}
	return lookup
}

func (r *Resolver) lookupGeoLite2(ip string) (region string, ok bool) {
	parsed := net.ParseIP(ip)
	if parsed == nil || IsLocalIP(ip) {
		return "", false
	}

	record, err := r.db.Country(parsed)
	if err != nil {
		return "", false
	}

	code := strings.ToUpper(record.Country.IsoCode)
	if !IsValidCountryCode(code) {
		return "", false
	}
	return code, true
}

// LookupCountry keeps GeoLite2-only lookup for diagnostic scripts.
func (r *Resolver) LookupCountry(ip string) (region string, ok bool) {
	return r.lookupGeoLite2(ip)
}
