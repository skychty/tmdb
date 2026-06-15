package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/sync/singleflight"

	"tmdb/internal/model"
	tmdbclient "tmdb/internal/tmdb"
)

type EnrichmentService struct {
	store     *CacheStore
	tmdb      *tmdbclient.Client
	imageBase string
	group     singleflight.Group
}

func NewEnrichmentService(store *CacheStore, tmdb *tmdbclient.Client, imageBase string) *EnrichmentService {
	return &EnrichmentService{
		store:     store,
		tmdb:      tmdb,
		imageBase: imageBase,
	}
}

func (s *EnrichmentService) EnrichMovies(ctx context.Context, language string, movies []model.Movie) []model.Movie {
	if len(movies) == 0 {
		return movies
	}

	var wg sync.WaitGroup
	for i := range movies {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			urls := s.movieMediaURLs(ctx, movies[index].ID, language)
			movies[index].TrailerURL = urls.TrailerURL
			movies[index].LogoURL = urls.LogoURL
		}(i)
	}
	wg.Wait()
	return movies
}

func (s *EnrichmentService) EnrichTVShows(ctx context.Context, language string, shows []model.TVShow) []model.TVShow {
	if len(shows) == 0 {
		return shows
	}

	var wg sync.WaitGroup
	for i := range shows {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			urls := s.tvMediaURLs(ctx, shows[index].ID, language)
			shows[index].TrailerURL = urls.TrailerURL
			shows[index].LogoURL = urls.LogoURL
		}(i)
	}
	wg.Wait()
	return shows
}

func (s *EnrichmentService) movieMediaURLs(ctx context.Context, movieID int, language string) model.MediaURLs {
	return s.lookupMediaURLs(ctx, "movie", movieID, language, func(ctx context.Context) ([]model.TMDBVideo, []model.TMDBLogo, error) {
		return s.tmdb.MovieMediaExtras(ctx, movieID, language)
	})
}

func (s *EnrichmentService) tvMediaURLs(ctx context.Context, tvID int, language string) model.MediaURLs {
	return s.lookupMediaURLs(ctx, "tv", tvID, language, func(ctx context.Context) ([]model.TMDBVideo, []model.TMDBLogo, error) {
		return s.tmdb.TVMediaExtras(ctx, tvID, language)
	})
}

type fetchMediaExtrasFunc func(ctx context.Context) ([]model.TMDBVideo, []model.TMDBLogo, error)

func (s *EnrichmentService) lookupMediaURLs(
	ctx context.Context,
	mediaType string,
	id int,
	language string,
	fetch fetchMediaExtrasFunc,
) model.MediaURLs {
	language = strings.TrimSpace(language)
	if language == "" {
		language = "en-US"
	}

	cacheKey := fmt.Sprintf("tmdb:media:v2:%s:%d:%s", mediaType, id, language)
	if cached, ok := s.loadMediaURLs(ctx, cacheKey); ok {
		return cached
	}

	val, err, _ := s.group.Do(cacheKey, func() (any, error) {
		if cached, ok := s.loadMediaURLs(ctx, cacheKey); ok {
			return cached, nil
		}

		videos, logos, err := fetch(ctx)
		if err != nil {
			return model.MediaURLs{}, err
		}

		urls := model.MediaURLs{
			TrailerURL: model.PickTrailerURL(videos),
			LogoURL:    model.PickLogoURL(logos, language, s.imageBase),
		}
		if data, err := json.Marshal(urls); err == nil {
			_ = s.store.Set(ctx, cacheKey, data)
		}
		return urls, nil
	})
	if err != nil {
		return model.MediaURLs{}
	}

	urls, _ := val.(model.MediaURLs)
	return urls
}

func (s *EnrichmentService) loadMediaURLs(ctx context.Context, key string) (model.MediaURLs, bool) {
	data, ok, err := s.store.GetFresh(ctx, key)
	if err != nil || !ok {
		return model.MediaURLs{}, false
	}
	var urls model.MediaURLs
	if err := json.Unmarshal(data, &urls); err != nil {
		return model.MediaURLs{}, false
	}
	return urls, true
}
