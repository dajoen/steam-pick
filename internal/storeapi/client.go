package storeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dajoen/steam-pick/internal/model"
)

const maxRetries = 2

// Client is the Steam Store API client.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Store API client.
func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i) * 500 * time.Millisecond)
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode >= 500 {
			_ = resp.Body.Close()
			continue
		}

		return resp, nil
	}

	return resp, err
}

// IsTurnBased checks if a game is turn-based by looking at genres and categories.
func (c *Client) IsTurnBased(ctx context.Context, appID int, country string) (bool, error) {
	u, _ := url.Parse("https://store.steampowered.com/api/appdetails")
	q := u.Query()
	q.Set("appids", fmt.Sprintf("%d", appID))
	q.Set("cc", country)
	q.Set("l", "english")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return false, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("store api returned status: %d", resp.StatusCode)
	}

	var result model.AppDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	appData, ok := result[fmt.Sprintf("%d", appID)]
	if !ok || !appData.Success {
		// If success is false, it might be region locked or invalid.
		// We treat it as not turn-based (or error).
		return false, fmt.Errorf("failed to get app details for %d", appID)
	}

	// Check genres and categories
	for _, g := range appData.Data.Genres {
		if strings.Contains(strings.ToLower(g.Description), "turn") {
			return true, nil
		}
	}
	for _, cat := range appData.Data.Categories {
		if strings.Contains(strings.ToLower(cat.Description), "turn") {
			return true, nil
		}
	}

	return false, nil
}
