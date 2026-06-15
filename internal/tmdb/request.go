package tmdb

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func newJSONRequest(ctx context.Context, endpoint, accessToken string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	return req, nil
}

func doRequest(client *http.Client, req *http.Request) ([]byte, int, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func decodeJSON(body []byte, target any) error {
	return json.Unmarshal(body, target)
}
