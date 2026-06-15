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

const trailerFallbackLanguage = "en-US"

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
	return s.lookupMediaURLs(ctx, "movie", movieID, language,
		func(ctx context.Context) ([]model.TMDBVideo, []model.TMDBLogo, error) {
			return s.tmdb.MovieMediaExtras(ctx, movieID, language)
		},
		func(ctx context.Context) ([]model.TMDBVideo, error) {
			resp, err := s.tmdb.MovieVideos(ctx, movieID, trailerFallbackLanguage)
			if err != nil {
				return nil, err
			}
			return resp.Results, nil
		},
	)
}

func (s *EnrichmentService) tvMediaURLs(ctx context.Context, tvID int, language string) model.MediaURLs {
	return s.lookupMediaURLs(ctx, "tv", tvID, language,
		func(ctx context.Context) ([]model.TMDBVideo, []model.TMDBLogo, error) {
			return s.tmdb.TVMediaExtras(ctx, tvID, language)
		},
		func(ctx context.Context) ([]model.TMDBVideo, error) {
			resp, err := s.tmdb.TVVideos(ctx, tvID, trailerFallbackLanguage)
			if err != nil {
				return nil, err
			}
			return resp.Results, nil
		},
	)
}

type fetchMediaExtrasFunc func(ctx context.Context) ([]model.TMDBVideo, []model.TMDBLogo, error)
type fetchEnUSVideosFunc func(ctx context.Context) ([]model.TMDBVideo, error)

func (s *EnrichmentService) lookupMediaURLs(
	ctx context.Context,
	mediaType string,
	id int,
	language string,
	fetch fetchMediaExtrasFunc,
	fetchEnUSVideos fetchEnUSVideosFunc,
) model.MediaURLs {
	language = strings.TrimSpace(language)
	if language == "" {
		language = trailerFallbackLanguage
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

		trailerURL := model.PickTrailerURL(videos)
		if trailerURL == "" && needsTrailerFallback(language) {
			trailerURL = s.trailerURLWithEnUSFallback(ctx, mediaType, id, fetchEnUSVideos)
		} else if isEnglishLanguage(language) {
			s.saveTrailerSubCache(ctx, mediaType, id, trailerURL)
		}

		urls := model.MediaURLs{
			TrailerURL: trailerURL,
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

func (s *EnrichmentService) trailerURLWithEnUSFallback(
	ctx context.Context,
	mediaType string,
	id int,
	fetchEnUSVideos fetchEnUSVideosFunc,
) string {
	subKey := buildTrailerSubCacheKey(mediaType, id)
	if cached, ok := s.loadTrailerSubCache(ctx, subKey); ok {
		return cached
	}

	flightKey := fmt.Sprintf("trailer:en-US:%s:%d", mediaType, id)
	val, err, _ := s.group.Do(flightKey, func() (any, error) {
		if cached, ok := s.loadTrailerSubCache(ctx, subKey); ok {
			return cached, nil
		}

		videos, err := fetchEnUSVideos(ctx)
		if err != nil {
			return "", err
		}

		trailerURL := model.PickTrailerURL(videos)
		s.saveTrailerSubCache(ctx, mediaType, id, trailerURL)
		return trailerURL, nil
	})
	if err != nil {
		return ""
	}

	trailerURL, _ := val.(string)
	return trailerURL
}

func buildTrailerSubCacheKey(mediaType string, id int) string {
	return fmt.Sprintf("tmdb:media:videos:v1:%s:%d:%s", mediaType, id, trailerFallbackLanguage)
}

func needsTrailerFallback(language string) bool {
	return !isEnglishLanguage(language)
}

func isEnglishLanguage(language string) bool {
	lang := strings.ToLower(strings.TrimSpace(language))
	return lang == "en-us" || lang == "en"
}

func (s *EnrichmentService) saveTrailerSubCache(ctx context.Context, mediaType string, id int, trailerURL string) {
	entry := model.MediaURLs{TrailerURL: trailerURL}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	_ = s.store.Set(ctx, buildTrailerSubCacheKey(mediaType, id), data)
}

func (s *EnrichmentService) loadTrailerSubCache(ctx context.Context, key string) (string, bool) {
	data, ok, err := s.store.GetFresh(ctx, key)
	if err != nil || !ok {
		return "", false
	}
	var entry model.MediaURLs
	if err := json.Unmarshal(data, &entry); err != nil {
		return "", false
	}
	return entry.TrailerURL, true
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
