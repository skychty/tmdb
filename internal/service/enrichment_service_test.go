package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"tmdb/internal/cache"
	"tmdb/internal/model"
)

type memoryCache struct {
	mu   sync.Mutex
	data map[string][]byte
}

func newMemoryCache() *memoryCache {
	return &memoryCache{data: make(map[string][]byte)}
}

func (m *memoryCache) Get(_ context.Context, key string) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, ok := m.data[key]
	return value, ok, nil
}

func (m *memoryCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

var _ cache.Cache = (*memoryCache)(nil)

func zhLogo() []model.TMDBLogo {
	zh := "zh"
	return []model.TMDBLogo{{FilePath: "/zh.png", Iso639_1: &zh, VoteAverage: 1, FileType: "png"}}
}

func enTrailerVideo() []model.TMDBVideo {
	return []model.TMDBVideo{{Key: "en-trailer", Site: "YouTube", Type: "Trailer", Official: true, Size: 1080}}
}

func TestLookupMediaURLsFallsBackToEnUSVideos(t *testing.T) {
	store := NewCacheStore(newMemoryCache(), time.Hour, time.Hour)
	svc := NewEnrichmentService(store, nil, "https://image.tmdb.org/t/p")

	enUSFetchCount := 0
	urls := svc.lookupMediaURLs(
		context.Background(),
		"movie",
		931285,
		"zh-CN",
		func(context.Context) ([]model.TMDBVideo, []model.TMDBLogo, error) {
			return nil, zhLogo(), nil
		},
		func(context.Context) ([]model.TMDBVideo, error) {
			enUSFetchCount++
			return enTrailerVideo(), nil
		},
	)

	wantTrailer := "https://www.youtube.com/watch?v=en-trailer"
	if urls.TrailerURL != wantTrailer {
		t.Fatalf("TrailerURL = %q, want %q", urls.TrailerURL, wantTrailer)
	}
	if urls.LogoURL != "https://image.tmdb.org/t/p/w185/zh.png" {
		t.Fatalf("LogoURL = %q, want zh logo", urls.LogoURL)
	}
	if enUSFetchCount != 1 {
		t.Fatalf("en-US fetch count = %d, want 1", enUSFetchCount)
	}
}

func TestLookupMediaURLsReusesEnUSVideosSubCacheAcrossLanguages(t *testing.T) {
	store := NewCacheStore(newMemoryCache(), time.Hour, time.Hour)
	svc := NewEnrichmentService(store, nil, "https://image.tmdb.org/t/p")

	enUSFetchCount := 0
	fetchEnUS := func(context.Context) ([]model.TMDBVideo, error) {
		enUSFetchCount++
		return enTrailerVideo(), nil
	}
	fetchExtras := func(context.Context) ([]model.TMDBVideo, []model.TMDBLogo, error) {
		return nil, zhLogo(), nil
	}

	for _, language := range []string{"zh-CN", "ja-JP"} {
		urls := svc.lookupMediaURLs(context.Background(), "movie", 931285, language, fetchExtras, fetchEnUS)
		if urls.TrailerURL != "https://www.youtube.com/watch?v=en-trailer" {
			t.Fatalf("language=%s TrailerURL = %q, want en trailer", language, urls.TrailerURL)
		}
	}

	if enUSFetchCount != 1 {
		t.Fatalf("en-US fetch count = %d, want 1", enUSFetchCount)
	}
}

func TestLookupMediaURLSWarmsEnUSVideosSubCache(t *testing.T) {
	store := NewCacheStore(newMemoryCache(), time.Hour, time.Hour)
	svc := NewEnrichmentService(store, nil, "https://image.tmdb.org/t/p")

	enUSFetchCount := 0
	urls := svc.lookupMediaURLs(
		context.Background(),
		"movie",
		931285,
		"en-US",
		func(context.Context) ([]model.TMDBVideo, []model.TMDBLogo, error) {
			return enTrailerVideo(), nil, nil
		},
		func(context.Context) ([]model.TMDBVideo, error) {
			enUSFetchCount++
			return enTrailerVideo(), nil
		},
	)
	if urls.TrailerURL != "https://www.youtube.com/watch?v=en-trailer" {
		t.Fatalf("TrailerURL = %q, want en trailer", urls.TrailerURL)
	}

	urls = svc.lookupMediaURLs(
		context.Background(),
		"movie",
		931285,
		"zh-CN",
		func(context.Context) ([]model.TMDBVideo, []model.TMDBLogo, error) {
			return nil, zhLogo(), nil
		},
		func(context.Context) ([]model.TMDBVideo, error) {
			enUSFetchCount++
			return enTrailerVideo(), nil
		},
	)
	if urls.TrailerURL != "https://www.youtube.com/watch?v=en-trailer" {
		t.Fatalf("zh-CN TrailerURL = %q, want warmed en trailer", urls.TrailerURL)
	}
	if enUSFetchCount != 0 {
		t.Fatalf("en-US fetch count after warm = %d, want 0", enUSFetchCount)
	}
}

func TestLookupMediaURLsSkipsFallbackForEnglish(t *testing.T) {
	store := NewCacheStore(newMemoryCache(), time.Hour, time.Hour)
	svc := NewEnrichmentService(store, nil, "https://image.tmdb.org/t/p")

	enUSFetchCount := 0
	urls := svc.lookupMediaURLs(
		context.Background(),
		"movie",
		931285,
		"en-US",
		func(context.Context) ([]model.TMDBVideo, []model.TMDBLogo, error) {
			return nil, nil, nil
		},
		func(context.Context) ([]model.TMDBVideo, error) {
			enUSFetchCount++
			return enTrailerVideo(), nil
		},
	)

	if urls.TrailerURL != "" {
		t.Fatalf("TrailerURL = %q, want empty without fallback fetch", urls.TrailerURL)
	}
	if enUSFetchCount != 0 {
		t.Fatalf("en-US fetch count = %d, want 0", enUSFetchCount)
	}
}
