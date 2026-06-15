package tmdb

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"tmdb/internal/model"
)

func (c *Client) MovieVideos(ctx context.Context, movieID int, language string) (model.TMDBVideosResponse, error) {
	path := fmt.Sprintf("/movie/%d/videos", movieID)
	return c.fetchVideos(ctx, path, language)
}

func (c *Client) TVVideos(ctx context.Context, tvID int, language string) (model.TMDBVideosResponse, error) {
	path := fmt.Sprintf("/tv/%d/videos", tvID)
	return c.fetchVideos(ctx, path, language)
}

func (c *Client) fetchVideos(ctx context.Context, path, language string) (model.TMDBVideosResponse, error) {
	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return model.TMDBVideosResponse{}, err
	}

	query := endpoint.Query()
	query.Set("language", language)
	if c.accessToken == "" && c.apiKey != "" {
		query.Set("api_key", c.apiKey)
	}
	endpoint.RawQuery = query.Encode()

	return fetchJSON[model.TMDBVideosResponse](c, ctx, endpoint.String())
}

func fetchJSON[T any](c *Client, ctx context.Context, endpoint string) (T, error) {
	var zero T
	if err := c.limiter.Acquire(ctx); err != nil {
		return zero, err
	}

	req, err := newJSONRequest(ctx, endpoint, c.accessToken)
	if err != nil {
		return zero, err
	}

	body, status, err := doRequest(c.httpClient, req)
	if err != nil {
		return zero, err
	}
	if status != 200 {
		return zero, fmt.Errorf("tmdb returned status %d: %s", status, string(body))
	}

	var result T
	if err := decodeJSON(body, &result); err != nil {
		return zero, err
	}
	return result, nil
}
