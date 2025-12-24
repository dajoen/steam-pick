package model

// Game represents a Steam game owned by a user.
type Game struct {
	AppID           int    `json:"appid"`
	Name            string `json:"name"`
	PlaytimeForever int    `json:"playtime_forever"`
	ImgIconURL      string `json:"img_icon_url"`
	// Store details (populated later)
	IsTurnBased bool   `json:"is_turn_based,omitempty"`
	StoreURL    string `json:"store_url,omitempty"`
}

// SteamResponse is the top-level response from GetOwnedGames.
type SteamResponse struct {
	Response struct {
		GameCount int    `json:"game_count"`
		Games     []Game `json:"games"`
	} `json:"response"`
}

// VanityResponse is the top-level response from ResolveVanityURL.
type VanityResponse struct {
	Response struct {
		SteamID string `json:"steamid"`
		Success int    `json:"success"`
		Message string `json:"message"`
	} `json:"response"`
}

// AppDetailsResponse is the top-level response from the Store API.
type AppDetailsResponse map[string]struct {
	Success bool `json:"success"`
	Data    struct {
		Name       string `json:"name"`
		Categories []struct {
			ID          int    `json:"id"`
			Description string `json:"description"`
		} `json:"categories"`
		Genres []struct {
			ID          string `json:"id"`
			Description string `json:"description"`
		} `json:"genres"`
	} `json:"data"`
}
