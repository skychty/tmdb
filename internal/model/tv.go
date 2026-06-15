package model

import (
	"strings"
	"time"
)

type TVListResponse struct {
	Page         int       `json:"page"`
	TotalPages   int       `json:"total_pages"`
	TotalResults int       `json:"total_results"`
	Region       string    `json:"region"`
	CachedAt     time.Time `json:"cached_at"`
	Results      []TVShow  `json:"results"`
}

type TVShow struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	OriginalName  string   `json:"original_name"`
	Overview      string   `json:"overview"`
	FirstAirDate  string   `json:"first_air_date"`
	PosterURL     string   `json:"poster_url"`
	BackdropURL   string   `json:"backdrop_url"`
	VoteAverage   float64  `json:"vote_average"`
	VoteCount     int      `json:"vote_count"`
	Popularity    float64  `json:"popularity"`
	GenreIDs      []int    `json:"genre_ids"`
	OriginCountry []string `json:"origin_country"`
	TrailerURL    string   `json:"trailer_url"`
	LogoURL       string   `json:"logo_url"`
}

type TMDBTVListResponse struct {
	Page         int       `json:"page"`
	TotalPages   int       `json:"total_pages"`
	TotalResults int       `json:"total_results"`
	Results      []TMDBTV  `json:"results"`
}

type TMDBTV struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	OriginalName  string   `json:"original_name"`
	Overview      string   `json:"overview"`
	FirstAirDate  string   `json:"first_air_date"`
	PosterPath    string   `json:"poster_path"`
	BackdropPath  string   `json:"backdrop_path"`
	VoteAverage   float64  `json:"vote_average"`
	VoteCount     int      `json:"vote_count"`
	Popularity    float64  `json:"popularity"`
	GenreIDs      []int    `json:"genre_ids"`
	OriginCountry []string `json:"origin_country"`
}

func ToTVListResponse(raw TMDBTVListResponse, region, imageBase string) TVListResponse {
	shows := make([]TVShow, 0, len(raw.Results))
	for _, item := range raw.Results {
		shows = append(shows, ToTVShow(item, imageBase))
	}

	return TVListResponse{
		Page:         raw.Page,
		TotalPages:   raw.TotalPages,
		TotalResults: raw.TotalResults,
		Region:       strings.ToUpper(region),
		CachedAt:     time.Now().UTC(),
		Results:      shows,
	}
}

func ToTVShow(item TMDBTV, imageBase string) TVShow {
	return TVShow{
		ID:            item.ID,
		Name:          item.Name,
		OriginalName:  item.OriginalName,
		Overview:      item.Overview,
		FirstAirDate:  item.FirstAirDate,
		PosterURL:     buildImageURL(imageBase, "w300", item.PosterPath),
		BackdropURL:   buildImageURL(imageBase, "original", item.BackdropPath),
		VoteAverage:   item.VoteAverage,
		VoteCount:     item.VoteCount,
		Popularity:    item.Popularity,
		GenreIDs:      item.GenreIDs,
		OriginCountry: item.OriginCountry,
	}
}
