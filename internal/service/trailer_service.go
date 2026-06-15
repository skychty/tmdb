package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/sync/singleflight"

	"tmdb/internal/model"
	tmdbclient "tmdb/internal/tmdb"
)

type TrailerService struct {
	store *CacheStore
	tmdb  *tmdbclient.Client
	group singleflight.Group
}

func NewTrailerService(store *CacheStore, tmdb *tmdbclient.Client) *TrailerService {
	return &TrailerService{
		store: store,
		tmdb:  tmdb,
	}
}

func (s *TrailerService) EnrichMovies(ctx context.Context, language string, movies []model.Movie) []model.Movie {
	if len(movies) == 0 {
		return movies
	}

	var wg sync.WaitGroup
	for i := range movies {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			movies[index].TrailerURL = s.movieTrailerURL(ctx, movies[index].ID, language)
		}(i)
	}
	wg.Wait()
	return movies
}

func (s *TrailerService) EnrichTVShows(ctx context.Context, language string, shows []model.TVShow) []model.TVShow {
	if len(shows) == 0 {
		return shows
	}

	var wg sync.WaitGroup
	for i := range shows {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			shows[index].TrailerURL = s.tvTrailerURL(ctx, shows[index].ID, language)
		}(i)
	}
	wg.Wait()
	return shows
}

func (s *TrailerService) movieTrailerURL(ctx context.Context, movieID int, language string) string {
	return s.lookupTrailerURL(ctx, "movie", movieID, language, func(ctx context.Context) (model.TMDBVideosResponse, error) {
		return s.tmdb.MovieVideos(ctx, movieID, language)
	})
}

func (s *TrailerService) tvTrailerURL(ctx context.Context, tvID int, language string) string {
	return s.lookupTrailerURL(ctx, "tv", tvID, language, func(ctx context.Context) (model.TMDBVideosResponse, error) {
		return s.tmdb.TVVideos(ctx, tvID, language)
	})
}

type fetchVideosFunc func(ctx context.Context) (model.TMDBVideosResponse, error)

func (s *TrailerService) lookupTrailerURL(
	ctx context.Context,
	mediaType string,
	id int,
	language string,
	fetch fetchVideosFunc,
) string {
	language = strings.TrimSpace(language)
	if language == "" {
		language = "en-US"
	}

	cacheKey := fmt.Sprintf("tmdb:trailer:%s:%d:%s", mediaType, id, language)
	if cached, ok := s.loadTrailer(ctx, cacheKey); ok {
		return cached
	}

	val, err, _ := s.group.Do(cacheKey, func() (any, error) {
		if cached, ok := s.loadTrailer(ctx, cacheKey); ok {
			return cached, nil
		}

		raw, err := fetch(ctx)
		if err != nil {
			return "", err
		}

		trailerURL := model.PickTrailerURL(raw.Results)
		_ = s.store.Set(ctx, cacheKey, []byte(trailerURL))
		return trailerURL, nil
	})
	if err != nil {
		return ""
	}

	url, _ := val.(string)
	return url
}

func (s *TrailerService) loadTrailer(ctx context.Context, key string) (string, bool) {
	data, ok, err := s.store.GetFresh(ctx, key)
	if err != nil || !ok {
		return "", false
	}
	return string(data), true
}
