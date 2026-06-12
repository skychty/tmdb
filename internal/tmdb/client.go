package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"tmdb/internal/model"
)

type Client struct {
	baseURL     string
	accessToken string
	apiKey      string
	httpClient  *http.Client
}

func NewClient(baseURL, accessToken, apiKey string) *Client {
	return &Client{
		baseURL:     baseURL,
		accessToken: accessToken,
		apiKey:      apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) NowPlaying(ctx context.Context, region, language string, page int) (model.TMDBMovieListResponse, error) {
	return c.fetchMovieList(ctx, "/movie/now_playing", region, language, page)
}

func (c *Client) Popular(ctx context.Context, region, language string, page int) (model.TMDBMovieListResponse, error) {
	return c.fetchMovieList(ctx, "/movie/popular", region, language, page)
}

func (c *Client) DiscoverRegionalPopular(ctx context.Context, region, language string, page int) (model.TMDBMovieListResponse, error) {
	endpoint, err := url.Parse(c.baseURL + "/discover/movie")
	if err != nil {
		return model.TMDBMovieListResponse{}, err
	}

	now := time.Now().UTC()
	query := endpoint.Query()
	query.Set("region", region)
	query.Set("language", language)
	query.Set("page", strconv.Itoa(page))
	query.Set("sort_by", "popularity.desc")
	query.Set("include_adult", "false")
	query.Set("include_video", "false")
	query.Set("with_release_type", "2|3")
	query.Set("release_date.gte", now.AddDate(0, -3, 0).Format("2006-01-02"))
	query.Set("release_date.lte", now.AddDate(0, 0, 30).Format("2006-01-02"))
	if c.accessToken == "" && c.apiKey != "" {
		query.Set("api_key", c.apiKey)
	}
	endpoint.RawQuery = query.Encode()

	return c.fetchFromURL(ctx, endpoint.String())
}

func (c *Client) fetchMovieList(ctx context.Context, path, region, language string, page int) (model.TMDBMovieListResponse, error) {
	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return model.TMDBMovieListResponse{}, err
	}

	query := endpoint.Query()
	query.Set("region", region)
	query.Set("language", language)
	query.Set("page", strconv.Itoa(page))
	if c.accessToken == "" && c.apiKey != "" {
		query.Set("api_key", c.apiKey)
	}
	endpoint.RawQuery = query.Encode()

	return c.fetchFromURL(ctx, endpoint.String())
}

func (c *Client) fetchFromURL(ctx context.Context, endpoint string) (model.TMDBMovieListResponse, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return model.TMDBMovieListResponse{}, err
		}
		req.Header.Set("Accept", "application/json")
		if c.accessToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.accessToken)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("tmdb returned status %d", resp.StatusCode)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return model.TMDBMovieListResponse{}, fmt.Errorf("tmdb returned status %d: %s", resp.StatusCode, string(body))
		}

		var result model.TMDBMovieListResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return model.TMDBMovieListResponse{}, err
		}
		return result, nil
	}

	return model.TMDBMovieListResponse{}, fmt.Errorf("tmdb request failed: %w", lastErr)
}

func (c *Client) OnTheAir(ctx context.Context, region, language string, page int) (model.TMDBTVListResponse, error) {
	return c.fetchTVList(ctx, "/tv/on_the_air", region, language, page)
}

func (c *Client) TVPopular(ctx context.Context, region, language string, page int) (model.TMDBTVListResponse, error) {
	return c.fetchTVList(ctx, "/tv/popular", region, language, page)
}

func (c *Client) DiscoverRegionalPopularTV(ctx context.Context, region, language string, page int) (model.TMDBTVListResponse, error) {
	endpoint, err := url.Parse(c.baseURL + "/discover/tv")
	if err != nil {
		return model.TMDBTVListResponse{}, err
	}

	now := time.Now().UTC()
	query := endpoint.Query()
	query.Set("with_origin_country", region)
	query.Set("language", language)
	query.Set("page", strconv.Itoa(page))
	query.Set("sort_by", "popularity.desc")
	query.Set("include_adult", "false")
	query.Set("first_air_date.gte", now.AddDate(0, -3, 0).Format("2006-01-02"))
	query.Set("first_air_date.lte", now.AddDate(0, 0, 30).Format("2006-01-02"))
	if c.accessToken == "" && c.apiKey != "" {
		query.Set("api_key", c.apiKey)
	}
	endpoint.RawQuery = query.Encode()

	return c.fetchTVFromURL(ctx, endpoint.String())
}

func (c *Client) fetchTVList(ctx context.Context, path, region, language string, page int) (model.TMDBTVListResponse, error) {
	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return model.TMDBTVListResponse{}, err
	}

	query := endpoint.Query()
	query.Set("language", language)
	query.Set("page", strconv.Itoa(page))
	if c.accessToken == "" && c.apiKey != "" {
		query.Set("api_key", c.apiKey)
	}
	endpoint.RawQuery = query.Encode()

	return c.fetchTVFromURL(ctx, endpoint.String())
}

func (c *Client) fetchTVFromURL(ctx context.Context, endpoint string) (model.TMDBTVListResponse, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return model.TMDBTVListResponse{}, err
		}
		req.Header.Set("Accept", "application/json")
		if c.accessToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.accessToken)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("tmdb returned status %d", resp.StatusCode)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return model.TMDBTVListResponse{}, fmt.Errorf("tmdb returned status %d: %s", resp.StatusCode, string(body))
		}

		var result model.TMDBTVListResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return model.TMDBTVListResponse{}, err
		}
		return result, nil
	}

	return model.TMDBTVListResponse{}, fmt.Errorf("tmdb request failed: %w", lastErr)
}
