package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/sync/singleflight"

	"tmdb/internal/model"
	tmdbclient "tmdb/internal/tmdb"
)

type MovieService struct {
	store     *CacheStore
	tmdb      *tmdbclient.Client
	trailers  *TrailerService
	imageBase string
	group     singleflight.Group
}

func NewMovieService(store *CacheStore, tmdb *tmdbclient.Client, trailers *TrailerService, imageBase string) *MovieService {
	return &MovieService{
		store:     store,
		tmdb:      tmdb,
		trailers:  trailers,
		imageBase: imageBase,
	}
}

func (s *MovieService) GetLatestMovies(ctx context.Context, region, language string, page, limit int) (model.MovieListResponse, error) {
	return s.getMovies(ctx, "latest", region, language, page, limit, s.tmdb.NowPlaying)
}

func (s *MovieService) GetPopularMovies(ctx context.Context, region, language string, page, limit int) (model.MovieListResponse, error) {
	return s.getMovies(ctx, "popular", region, language, page, limit, s.tmdb.Popular)
}

func (s *MovieService) GetRegionalPopularMovies(ctx context.Context, region, language string, page, limit int) (model.MovieListResponse, error) {
	return s.getMovies(ctx, "regional-popular", region, language, page, limit, s.tmdb.DiscoverRegionalPopular)
}

type fetchFunc func(ctx context.Context, region, language string, page int) (model.TMDBMovieListResponse, error)

func (s *MovieService) getMovies(
	ctx context.Context,
	listType, region, language string,
	page, limit int,
	fetch fetchFunc,
) (model.MovieListResponse, error) {
	region = strings.ToUpper(strings.TrimSpace(region))
	language = strings.TrimSpace(language)
	if language == "" {
		language = "en-US"
	}
	limit = normalizePageLimit(limit)

	cacheKey := buildCacheKey(listType, region, language, page)
	if resp, ok := s.loadFreshMovie(ctx, cacheKey); ok {
		return s.finalizeMovieResponse(ctx, resp, language, limit), nil
	}

	val, err, _ := s.group.Do(cacheKey, func() (any, error) {
		if resp, ok := s.loadFreshMovie(ctx, cacheKey); ok {
			return resp, nil
		}

		raw, err := fetch(ctx, region, language, page)
		if err != nil {
			return s.loadStaleMovieRaw(ctx, cacheKey, err)
		}

		resp := model.ToMovieListResponse(raw, region, s.imageBase)
		if data, err := json.Marshal(resp); err == nil {
			_ = s.store.Set(ctx, cacheKey, data)
		}
		return resp, nil
	})
	if err != nil {
		return model.MovieListResponse{}, err
	}

	resp, ok := val.(model.MovieListResponse)
	if !ok {
		return model.MovieListResponse{}, fmt.Errorf("unexpected cache value type")
	}
	return s.finalizeMovieResponse(ctx, resp, language, limit), nil
}

func (s *MovieService) finalizeMovieResponse(ctx context.Context, resp model.MovieListResponse, language string, limit int) model.MovieListResponse {
	limit = normalizePageLimit(limit)
	if len(resp.Results) > limit {
		resp.Results = resp.Results[:limit]
	}
	resp.Results = s.trailers.EnrichMovies(ctx, language, resp.Results)
	return resp
}

func (s *MovieService) loadFreshMovie(ctx context.Context, key string) (model.MovieListResponse, bool) {
	data, ok, err := s.store.GetFresh(ctx, key)
	if err != nil || !ok {
		return model.MovieListResponse{}, false
	}
	var resp model.MovieListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return model.MovieListResponse{}, false
	}
	return resp, true
}

func (s *MovieService) loadStaleMovieRaw(ctx context.Context, key string, cause error) (model.MovieListResponse, error) {
	data, ok, err := s.store.GetStale(ctx, key)
	if err != nil {
		return model.MovieListResponse{}, cause
	}
	if !ok {
		return model.MovieListResponse{}, cause
	}
	var resp model.MovieListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return model.MovieListResponse{}, cause
	}
	return resp, nil
}

func buildCacheKey(listType, region, language string, page int) string {
	return fmt.Sprintf("tmdb:movies:%s:%s:%s:%d", listType, region, language, page)
}
