package model

import (
	"fmt"
	"strings"
	"time"
)

type MovieListResponse struct {
	Page         int       `json:"page"`
	TotalPages   int       `json:"total_pages"`
	TotalResults int       `json:"total_results"`
	Region       string    `json:"region"`
	CachedAt     time.Time `json:"cached_at"`
	Results      []Movie   `json:"results"`
}

type Movie struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	Overview      string  `json:"overview"`
	ReleaseDate   string  `json:"release_date"`
	PosterURL     string  `json:"poster_url"`
	BackdropURL   string  `json:"backdrop_url"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int     `json:"vote_count"`
	Popularity    float64 `json:"popularity"`
	GenreIDs      []int   `json:"genre_ids"`
}

type TMDBMovieListResponse struct {
	Page         int         `json:"page"`
	TotalPages   int         `json:"total_pages"`
	TotalResults int         `json:"total_results"`
	Results      []TMDBMovie `json:"results"`
}

type TMDBMovie struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	Overview      string  `json:"overview"`
	ReleaseDate   string  `json:"release_date"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int     `json:"vote_count"`
	Popularity    float64 `json:"popularity"`
	GenreIDs      []int   `json:"genre_ids"`
}

func ToMovieListResponse(raw TMDBMovieListResponse, region, imageBase string) MovieListResponse {
	movies := make([]Movie, 0, len(raw.Results))
	for _, item := range raw.Results {
		movies = append(movies, ToMovie(item, imageBase))
	}

	return MovieListResponse{
		Page:         raw.Page,
		TotalPages:   raw.TotalPages,
		TotalResults: raw.TotalResults,
		Region:       strings.ToUpper(region),
		CachedAt:     time.Now().UTC(),
		Results:      movies,
	}
}

func ToMovie(item TMDBMovie, imageBase string) Movie {
	return Movie{
		ID:            item.ID,
		Title:         item.Title,
		OriginalTitle: item.OriginalTitle,
		Overview:      item.Overview,
		ReleaseDate:   item.ReleaseDate,
		PosterURL:     buildImageURL(imageBase, "w500", item.PosterPath),
		BackdropURL:   buildImageURL(imageBase, "original", item.BackdropPath),
		VoteAverage:   item.VoteAverage,
		VoteCount:     item.VoteCount,
		Popularity:    item.Popularity,
		GenreIDs:      item.GenreIDs,
	}
}

func buildImageURL(base, size, path string) string {
	if path == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s%s", strings.TrimRight(base, "/"), size, path)
}
