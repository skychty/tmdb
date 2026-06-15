package tmdb

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"tmdb/internal/model"
)

type mediaExtrasResponse struct {
	Videos struct {
		Results []model.TMDBVideo `json:"results"`
	} `json:"videos"`
	Images model.TMDBImagesResponse `json:"images"`
}

func (c *Client) MovieMediaExtras(ctx context.Context, movieID int, language string) ([]model.TMDBVideo, []model.TMDBLogo, error) {
	path := fmt.Sprintf("/movie/%d", movieID)
	return c.fetchMediaExtras(ctx, path, language)
}

func (c *Client) TVMediaExtras(ctx context.Context, tvID int, language string) ([]model.TMDBVideo, []model.TMDBLogo, error) {
	path := fmt.Sprintf("/tv/%d", tvID)
	return c.fetchMediaExtras(ctx, path, language)
}

func (c *Client) fetchMediaExtras(ctx context.Context, path, language string) ([]model.TMDBVideo, []model.TMDBLogo, error) {
	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, nil, err
	}

	langPrefix := languagePrefix(language)
	query := endpoint.Query()
	query.Set("language", language)
	query.Set("append_to_response", "videos,images")
	query.Set("include_image_language", langPrefix+",null,en")
	if c.accessToken == "" && c.apiKey != "" {
		query.Set("api_key", c.apiKey)
	}
	endpoint.RawQuery = query.Encode()

	raw, err := fetchJSON[mediaExtrasResponse](c, ctx, endpoint.String())
	if err != nil {
		return nil, nil, err
	}

	return raw.Videos.Results, raw.Images.Logos, nil
}

func languagePrefix(language string) string {
	language = strings.TrimSpace(language)
	if language == "" {
		return "en"
	}
	if idx := strings.Index(language, "-"); idx > 0 {
		return strings.ToLower(language[:idx])
	}
	return strings.ToLower(language)
}
