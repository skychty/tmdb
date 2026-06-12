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

type MovieService struct {
	cache     cache.Cache
	tmdb      *tmdbclient.Client
	imageBase string
	cacheTTL  time.Duration
	group     singleflight.Group
}

func NewMovieService(cache cache.Cache, tmdb *tmdbclient.Client, imageBase string, cacheTTL time.Duration) *MovieService {
	return &MovieService{
		cache:     cache,
		tmdb:      tmdb,
		imageBase: imageBase,
		cacheTTL:  cacheTTL,
	}
}

func (s *MovieService) GetLatestMovies(ctx context.Context, region, language string, page int) (model.MovieListResponse, error) {
	return s.getMovies(ctx, "latest", region, language, page, s.tmdb.NowPlaying)
}

func (s *MovieService) GetPopularMovies(ctx context.Context, region, language string, page int) (model.MovieListResponse, error) {
	return s.getMovies(ctx, "popular", region, language, page, s.tmdb.Popular)
}

type fetchFunc func(ctx context.Context, region, language string, page int) (model.TMDBMovieListResponse, error)

func (s *MovieService) getMovies(
	ctx context.Context,
	listType, region, language string,
	page int,
	fetch fetchFunc,
) (model.MovieListResponse, error) {
	region = strings.ToUpper(strings.TrimSpace(region))
	language = strings.TrimSpace(language)
	if language == "" {
		language = "zh-CN"
	}

	cacheKey := buildCacheKey(listType, region, language, page)
	if data, ok, err := s.cache.Get(ctx, cacheKey); err == nil && ok {
		var resp model.MovieListResponse
		if err := json.Unmarshal(data, &resp); err == nil {
			return resp, nil
		}
	}

	val, err, _ := s.group.Do(cacheKey, func() (any, error) {
		if data, ok, err := s.cache.Get(ctx, cacheKey); err == nil && ok {
			var resp model.MovieListResponse
			if err := json.Unmarshal(data, &resp); err == nil {
				return resp, nil
			}
		}

		raw, err := fetch(ctx, region, language, page)
		if err != nil {
			return model.MovieListResponse{}, err
		}

		resp := model.ToMovieListResponse(raw, region, s.imageBase)
		data, err := json.Marshal(resp)
		if err != nil {
			return resp, nil
		}
		_ = s.cache.Set(ctx, cacheKey, data, s.cacheTTL)
		return resp, nil
	})
	if err != nil {
		return model.MovieListResponse{}, err
	}

	resp, ok := val.(model.MovieListResponse)
	if !ok {
		return model.MovieListResponse{}, fmt.Errorf("unexpected cache value type")
	}
	return resp, nil
}

func buildCacheKey(listType, region, language string, page int) string {
	return fmt.Sprintf("tmdb:movies:%s:%s:%s:%d", listType, region, language, page)
}
