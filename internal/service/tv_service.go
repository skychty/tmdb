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

type TVService struct {
	store     *CacheStore
	tmdb      *tmdbclient.Client
	trailers  *TrailerService
	imageBase string
	group     singleflight.Group
}

func NewTVService(store *CacheStore, tmdb *tmdbclient.Client, trailers *TrailerService, imageBase string) *TVService {
	return &TVService{
		store:     store,
		tmdb:      tmdb,
		trailers:  trailers,
		imageBase: imageBase,
	}
}

func (s *TVService) GetOnTheAirTV(ctx context.Context, region, language string, page, limit int) (model.TVListResponse, error) {
	return s.getTVShows(ctx, "on-the-air", region, language, page, limit, s.tmdb.OnTheAir)
}

func (s *TVService) GetPopularTV(ctx context.Context, region, language string, page, limit int) (model.TVListResponse, error) {
	return s.getTVShows(ctx, "popular", region, language, page, limit, s.tmdb.TVPopular)
}

func (s *TVService) GetRegionalPopularTV(ctx context.Context, region, language string, page, limit int) (model.TVListResponse, error) {
	return s.getTVShows(ctx, "regional-popular-v2", region, language, page, limit, s.tmdb.DiscoverRegionalPopularTV)
}

type fetchTVFunc func(ctx context.Context, region, language string, page int) (model.TMDBTVListResponse, error)

func (s *TVService) getTVShows(
	ctx context.Context,
	listType, region, language string,
	page, limit int,
	fetch fetchTVFunc,
) (model.TVListResponse, error) {
	region = strings.ToUpper(strings.TrimSpace(region))
	language = strings.TrimSpace(language)
	if language == "" {
		language = "en-US"
	}
	limit = normalizePageLimit(limit)

	cacheKey := buildTVCacheKey(listType, region, language, page)
	if resp, ok := s.loadFreshTV(ctx, cacheKey); ok {
		return s.finalizeTVResponse(ctx, resp, language, limit), nil
	}

	val, err, _ := s.group.Do(cacheKey, func() (any, error) {
		if resp, ok := s.loadFreshTV(ctx, cacheKey); ok {
			return resp, nil
		}

		raw, err := fetch(ctx, region, language, page)
		if err != nil {
			return s.loadStaleTVRaw(ctx, cacheKey, err)
		}

		resp := model.ToTVListResponse(raw, region, s.imageBase)
		if data, err := json.Marshal(resp); err == nil {
			_ = s.store.Set(ctx, cacheKey, data)
		}
		return resp, nil
	})
	if err != nil {
		return model.TVListResponse{}, err
	}

	resp, ok := val.(model.TVListResponse)
	if !ok {
		return model.TVListResponse{}, fmt.Errorf("unexpected cache value type")
	}
	return s.finalizeTVResponse(ctx, resp, language, limit), nil
}

func (s *TVService) finalizeTVResponse(ctx context.Context, resp model.TVListResponse, language string, limit int) model.TVListResponse {
	limit = normalizePageLimit(limit)
	if len(resp.Results) > limit {
		resp.Results = resp.Results[:limit]
	}
	resp.Results = s.trailers.EnrichTVShows(ctx, language, resp.Results)
	return resp
}

func (s *TVService) loadFreshTV(ctx context.Context, key string) (model.TVListResponse, bool) {
	data, ok, err := s.store.GetFresh(ctx, key)
	if err != nil || !ok {
		return model.TVListResponse{}, false
	}
	var resp model.TVListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return model.TVListResponse{}, false
	}
	return resp, true
}

func (s *TVService) loadStaleTVRaw(ctx context.Context, key string, cause error) (model.TVListResponse, error) {
	data, ok, err := s.store.GetStale(ctx, key)
	if err != nil {
		return model.TVListResponse{}, cause
	}
	if !ok {
		return model.TVListResponse{}, cause
	}
	var resp model.TVListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return model.TVListResponse{}, cause
	}
	return resp, nil
}

func buildTVCacheKey(listType, region, language string, page int) string {
	return fmt.Sprintf("tmdb:tv:%s:%s:%s:%d", listType, region, language, page)
}
