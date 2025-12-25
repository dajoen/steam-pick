package pcgw

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

const (
	baseURL = "https://www.pcgamingwiki.com/w/api.php"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type CargoResponse struct {
	CargoQuery []struct {
		Title struct {
			SteamAppID string `json:"Steam_AppID"`
			Developers string `json:"Developers"`
			Publishers string `json:"Publishers"`
			Genres     string `json:"Genres"`
		} `json:"title"`
	} `json:"cargoquery"`
}

func (c *Client) GetAppDetails(ctx context.Context, appID int) (*model.AppDetailsResponse, error) {
	u, _ := url.Parse(baseURL)
	q := u.Query()
	q.Set("action", "cargoquery")
	q.Set("tables", "Infobox_game")
	q.Set("fields", "Steam_AppID,Developers,Publishers,Genres")
	q.Set("where", fmt.Sprintf("Steam_AppID HOLDS \"%d\"", appID))
	q.Set("format", "json")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set User-Agent as requested by MediaWiki API policy
	req.Header.Set("User-Agent", "steam-pick/1.0 (github.com/dajoen/steam-pick)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pcgw api returned status: %d", resp.StatusCode)
	}

	var result CargoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.CargoQuery) == 0 {
		return nil, fmt.Errorf("no results found for appid %d", appID)
	}

	data := result.CargoQuery[0].Title

	// Convert to model.AppDetailsResponse format
	// Note: PCGW genres are comma separated strings like "ARPG,"
	// We need to clean them up.

	genres := []model.Genre{}
	rawGenres := strings.Split(data.Genres, ",")
	for _, g := range rawGenres {
		g = strings.TrimSpace(g)
		if g != "" {
			genres = append(genres, model.Genre{Description: g})
		}
	}

	// Developers/Publishers in PCGW are often "Company:Name"
	// We don't have a field for Dev/Pub in our AppDetails struct yet (it's in the JSON but not fully typed in some versions)
	// But we do have Categories and Genres.

	// We'll construct a minimal AppDetails
	details := model.AppDetails{
		Name:                "Fetched from PCGamingWiki", // We don't get the name from this query easily, but we have it in DB
		DetailedDescription: "Data fetched from PCGamingWiki because Steam Store page is unavailable.",
		ShortDescription:    "Data fetched from PCGamingWiki.",
		Genres:              genres,
		Categories:          []model.Category{}, // PCGW doesn't map 1:1 to Steam Categories easily
	}

	response := make(model.AppDetailsResponse)
	response[fmt.Sprintf("%d", appID)] = model.AppDetailsEntry{
		Success: true,
		Data:    details,
	}

	return &response, nil
}
