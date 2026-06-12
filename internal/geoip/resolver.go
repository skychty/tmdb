package geoip

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"tmdb/internal/cache"
)

type Resolver struct {
	cache         cache.Cache
	defaultRegion string
	cacheTTL      time.Duration
	httpClient    *http.Client
}

func NewResolver(cache cache.Cache, defaultRegion string, cacheTTL time.Duration) *Resolver {
	return &Resolver{
		cache:         cache,
		defaultRegion: strings.ToUpper(defaultRegion),
		cacheTTL:      cacheTTL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (r *Resolver) ResolveRegion(ctx context.Context, ip string) (string, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsUnspecified() {
		return r.defaultRegion, nil
	}

	cacheKey := "geoip:region:" + ip
	if data, ok, err := r.cache.Get(ctx, cacheKey); err == nil && ok {
		return string(data), nil
	}

	region, err := r.lookupRegion(ctx, ip)
	if err != nil {
		return r.defaultRegion, nil
	}

	_ = r.cache.Set(ctx, cacheKey, []byte(region), r.cacheTTL)
	return region, nil
}

type ipAPIResponse struct {
	Status      string `json:"status"`
	CountryCode string `json:"countryCode"`
}

func (r *Resolver) lookupRegion(ctx context.Context, ip string) (string, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,countryCode", ip)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("geoip lookup returned status %d", resp.StatusCode)
	}

	var result ipAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if result.Status != "success" || len(result.CountryCode) != 2 {
		return "", fmt.Errorf("geoip lookup failed for ip %s", ip)
	}

	return strings.ToUpper(result.CountryCode), nil
}
