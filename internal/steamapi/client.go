package steamapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dajoen/steam-pick/internal/cache"
	"github.com/dajoen/steam-pick/internal/model"
)

var ErrRateLimitExceeded = fmt.Errorf("rate limit exceeded")

const (
	baseURL    = "https://api.steampowered.com"
	maxRetries = 2
)

// Client is the Steam Web API client.
type Client struct {
	apiKey      string
	httpClient  *http.Client
	gamesCache  *cache.Cache[model.SteamResponse]
	vanityCache *cache.Cache[model.VanityResponse]
	cacheTTL    time.Duration
}

// NewClient creates a new Steam API client.
func NewClient(apiKey string, ttl time.Duration, timeout time.Duration) (*Client, error) {
	gc, err := cache.New[model.SteamResponse]("steam-pick")
	if err != nil {
		return nil, fmt.Errorf("failed to init games cache: %w", err)
	}
	vc, err := cache.New[model.VanityResponse]("steam-pick")
	if err != nil {
		return nil, fmt.Errorf("failed to init vanity cache: %w", err)
	}

	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		gamesCache:  gc,
		vanityCache: vc,
		cacheTTL:    ttl,
	}, nil
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
			resp.Body.Close()
			continue
		}

		if resp.StatusCode == 429 {
			resp.Body.Close()
			return nil, ErrRateLimitExceeded
		}

		return resp, nil
	}

	return resp, err
}

// ResolveVanityURL resolves a vanity URL to a SteamID64.
func (c *Client) ResolveVanityURL(ctx context.Context, vanityURL string) (string, error) {
	cacheKey := "vanity_" + vanityURL
	if cached, found, err := c.vanityCache.Get(cacheKey, c.cacheTTL); err == nil && found {
		return cached.Response.SteamID, nil
	}

	u, _ := url.Parse(baseURL + "/ISteamUser/ResolveVanityURL/v1/")
	q := u.Query()
	q.Set("key", c.apiKey)
	q.Set("vanityurl", vanityURL)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("steam api returned status: %d", resp.StatusCode)
	}

	var result model.VanityResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Response.Success != 1 {
		return "", fmt.Errorf("vanity resolution failed: %s", result.Response.Message)
	}

	_ = c.vanityCache.Set(cacheKey, result)

	return result.Response.SteamID, nil
}

// GetOwnedGames returns the list of owned games.
func (c *Client) GetOwnedGames(ctx context.Context, steamID64 string, includeFree bool) ([]model.Game, error) {
	cacheKey := "owned_games_" + steamID64
	if includeFree {
		cacheKey += "_free"
	}

	if cached, found, err := c.gamesCache.Get(cacheKey, c.cacheTTL); err == nil && found {
		return cached.Response.Games, nil
	}

	u, _ := url.Parse(baseURL + "/IPlayerService/GetOwnedGames/v1/")
	q := u.Query()
	q.Set("key", c.apiKey)
	q.Set("steamid", steamID64)
	q.Set("include_appinfo", "1")
	q.Set("format", "json")
	if includeFree {
		q.Set("include_played_free_games", "1")
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam api returned status: %d", resp.StatusCode)
	}

	var result model.SteamResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if empty (private profile often returns empty games list but success)
	if len(result.Response.Games) == 0 && result.Response.GameCount == 0 {
		// This could be a private profile or just no games.
		// We can't easily distinguish without another API call, but we should warn the user.
		// For now, just return empty.
		return nil, nil
	}

	_ = c.gamesCache.Set(cacheKey, result)

	return result.Response.Games, nil
}

// GetAppDetails fetches store details for an app.
func (c *Client) GetAppDetails(ctx context.Context, appID int) (*model.AppDetailsResponse, error) {
	u, _ := url.Parse("https://store.steampowered.com/api/appdetails")
	q := u.Query()
	q.Set("appids", fmt.Sprintf("%d", appID))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam store api returned status: %d", resp.StatusCode)
	}

	var result model.AppDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
