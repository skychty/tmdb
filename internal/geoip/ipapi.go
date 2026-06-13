package geoip

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ipAPIResponse struct {
	Status      string `json:"status"`
	CountryCode string `json:"countryCode"`
}

func lookupIPAPI(ctx context.Context, ip string, client *http.Client) (string, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,countryCode", ip)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ip-api returned status %d", resp.StatusCode)
	}

	var result ipAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if result.Status != "success" || !IsValidCountryCode(result.CountryCode) {
		return "", fmt.Errorf("ip-api lookup failed for ip %s", ip)
	}

	return strings.ToUpper(result.CountryCode), nil
}

func newIPAPIClient() *http.Client {
	return &http.Client{Timeout: 5 * time.Second}
}
