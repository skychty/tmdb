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

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
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
