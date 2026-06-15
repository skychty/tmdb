package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tmdb/internal/api"
	"tmdb/internal/cache"
	"tmdb/internal/config"
	"tmdb/internal/geoip"
	"tmdb/internal/service"
	tmdbclient "tmdb/internal/tmdb"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	redisCache := cache.NewRedisCache(cfg.RedisAddr)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisCache.Ping(ctx); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}

	tmdbLimiter := tmdbclient.NewRateLimiter(cfg.TMDBRateLimit, cfg.TMDBRateBurst, cfg.TMDBQueueTimeout)
	tmdbClient := tmdbclient.NewClient(cfg.TMDBBaseURL, cfg.TMDBAccessToken, cfg.TMDBAPIKey, tmdbLimiter)
	cacheStore := service.NewCacheStore(redisCache, cfg.CacheTTL, cfg.StaleCacheTTL)
	trailerService := service.NewTrailerService(cacheStore, tmdbClient)
	movieService := service.NewMovieService(cacheStore, tmdbClient, trailerService, cfg.TMDBImageBase)
	tvService := service.NewTVService(cacheStore, tmdbClient, trailerService, cfg.TMDBImageBase)
	geoIPResolver, err := geoip.NewResolver(cfg.GeoIPDBPath, cfg.DefaultRegion, redisCache, cfg.GeoIPCacheTTL)
	if err != nil {
		log.Fatalf("geoip init: %v", err)
	}
	defer geoIPResolver.Close()

	router := api.NewRouter(movieService, tvService, geoIPResolver)

	addr := fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
	log.Printf("server listening on http://%s", addr)
		log.Printf("  tmdb rate limit: %.0f req/s, queue timeout: %s, stale cache: %s",
			cfg.TMDBRateLimit, cfg.TMDBQueueTimeout, cfg.StaleCacheTTL)
		log.Printf("  geoip db: %s, cache ttl: %s", cfg.GeoIPDBPath, cfg.GeoIPCacheTTL)
		for _, ip := range listAccessibleIPs(cfg.HTTPPort) {
			log.Printf("  accessible at http://%s", ip)
		}
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}

func listAccessibleIPs(port int) []string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}

	var result []string
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		ip := ipNet.IP.To4()
		if ip == nil {
			continue
		}
		result = append(result, fmt.Sprintf("%s:%d", ip.String(), port))
	}
	return result
}
