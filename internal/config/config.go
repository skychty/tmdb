package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	TMDBAccessToken string
	TMDBAPIKey      string
	TMDBBaseURL     string
	TMDBImageBase   string
	RedisAddr         string
	CacheTTL          time.Duration
	StaleCacheTTL     time.Duration
	TMDBRateLimit     float64
	TMDBRateBurst     int
	TMDBQueueTimeout  time.Duration
	HTTPHost          string
	HTTPPort        int
	DefaultRegion   string
	GeoIPCacheTTL   time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		TMDBAccessToken: os.Getenv("TMDB_ACCESS_TOKEN"),
		TMDBAPIKey:      os.Getenv("TMDB_API_KEY"),
		TMDBBaseURL:     getEnv("TMDB_BASE_URL", "https://api.themoviedb.org/3"),
		TMDBImageBase:   getEnv("TMDB_IMAGE_BASE", "https://image.tmdb.org/t/p"),
		RedisAddr:       getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		HTTPHost:        getEnv("HTTP_HOST", "0.0.0.0"),
		HTTPPort:        getEnvInt("HTTP_PORT", 8080),
		DefaultRegion:   strings.ToUpper(getEnv("DEFAULT_REGION", "CN")),
	}

	ttlStr := getEnv("CACHE_TTL", "24h")
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CACHE_TTL %q: %w", ttlStr, err)
	}
	cfg.CacheTTL = ttl

	staleTTLStr := getEnv("STALE_CACHE_TTL", "168h")
	staleTTL, err := time.ParseDuration(staleTTLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid STALE_CACHE_TTL %q: %w", staleTTLStr, err)
	}
	cfg.StaleCacheTTL = staleTTL

	queueTimeoutStr := getEnv("TMDB_QUEUE_TIMEOUT", "5s")
	queueTimeout, err := time.ParseDuration(queueTimeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid TMDB_QUEUE_TIMEOUT %q: %w", queueTimeoutStr, err)
	}
	cfg.TMDBQueueTimeout = queueTimeout

	cfg.TMDBRateLimit = getEnvFloat("TMDB_RATE_LIMIT", 40)
	cfg.TMDBRateBurst = getEnvInt("TMDB_RATE_BURST", int(cfg.TMDBRateLimit))

	geoIPCacheStr := getEnv("GEOIP_CACHE_TTL", "24h")
	geoIPCacheTTL, err := time.ParseDuration(geoIPCacheStr)
	if err != nil {
		return nil, fmt.Errorf("invalid GEOIP_CACHE_TTL %q: %w", geoIPCacheStr, err)
	}
	cfg.GeoIPCacheTTL = geoIPCacheTTL

	if cfg.TMDBAccessToken == "" && cfg.TMDBAPIKey == "" {
		return nil, fmt.Errorf("TMDB_ACCESS_TOKEN or TMDB_API_KEY is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return n
}
