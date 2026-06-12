package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"

	"tmdb/internal/cache"
	"tmdb/internal/model"
	tmdbclient "tmdb/internal/tmdb"
)

type TVService struct {
	cache     cache.Cache
	tmdb      *tmdbclient.Client
	imageBase string
	cacheTTL  time.Duration
	group     singleflight.Group
}

func NewTVService(cache cache.Cache, tmdb *tmdbclient.Client, imageBase string, cacheTTL time.Duration) *TVService {
	return &TVService{
		cache:     cache,
		tmdb:      tmdb,
		imageBase: imageBase,
		cacheTTL:  cacheTTL,
	}
}

func (s *TVService) GetOnTheAirTV(ctx context.Context, region, language string, page int) (model.TVListResponse, error) {
	return s.getTVShows(ctx, "on-the-air", region, language, page, s.tmdb.OnTheAir)
}

func (s *TVService) GetPopularTV(ctx context.Context, region, language string, page int) (model.TVListResponse, error) {
	return s.getTVShows(ctx, "popular", region, language, page, s.tmdb.TVPopular)
}

func (s *TVService) GetRegionalPopularTV(ctx context.Context, region, language string, page int) (model.TVListResponse, error) {
	return s.getTVShows(ctx, "regional-popular", region, language, page, s.tmdb.DiscoverRegionalPopularTV)
}

type fetchTVFunc func(ctx context.Context, region, language string, page int) (model.TMDBTVListResponse, error)

func (s *TVService) getTVShows(
	ctx context.Context,
	listType, region, language string,
	page int,
	fetch fetchTVFunc,
) (model.TVListResponse, error) {
	region = strings.ToUpper(strings.TrimSpace(region))
	language = strings.TrimSpace(language)
	if language == "" {
		language = "zh-CN"
	}

	cacheKey := buildTVCacheKey(listType, region, language, page)
	if data, ok, err := s.cache.Get(ctx, cacheKey); err == nil && ok {
		var resp model.TVListResponse
		if err := json.Unmarshal(data, &resp); err == nil {
			return resp, nil
		}
	}

	val, err, _ := s.group.Do(cacheKey, func() (any, error) {
		if data, ok, err := s.cache.Get(ctx, cacheKey); err == nil && ok {
			var resp model.TVListResponse
			if err := json.Unmarshal(data, &resp); err == nil {
				return resp, nil
			}
		}

		raw, err := fetch(ctx, region, language, page)
		if err != nil {
			return model.TVListResponse{}, err
		}

		resp := model.ToTVListResponse(raw, region, s.imageBase)
		data, err := json.Marshal(resp)
		if err != nil {
			return resp, nil
		}
		_ = s.cache.Set(ctx, cacheKey, data, s.cacheTTL)
		return resp, nil
	})
	if err != nil {
		return model.TVListResponse{}, err
	}

	resp, ok := val.(model.TVListResponse)
	if !ok {
		return model.TVListResponse{}, fmt.Errorf("unexpected cache value type")
	}
	return resp, nil
}

func buildTVCacheKey(listType, region, language string, page int) string {
	return fmt.Sprintf("tmdb:tv:%s:%s:%s:%d", listType, region, language, page)
}
