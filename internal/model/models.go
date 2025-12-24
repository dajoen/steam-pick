package model

// Game represents a Steam game owned by a user.
type Game struct {
	AppID                    int    `json:"appid"`
	Name                     string `json:"name"`
	PlaytimeForever          int    `json:"playtime_forever"`
	ImgIconURL               string `json:"img_icon_url"`
	HasCommunityVisibleStats bool   `json:"has_community_visible_stats,omitempty"`
	PlaytimeWindowsForever   int    `json:"playtime_windows_forever,omitempty"`
	PlaytimeMacForever       int    `json:"playtime_mac_forever,omitempty"`
	PlaytimeLinuxForever     int    `json:"playtime_linux_forever,omitempty"`
	PlaytimeDeckForever      int    `json:"playtime_deck_forever,omitempty"`
	RTimeLastPlayed          int    `json:"rtime_last_played,omitempty"`
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
type AppDetailsResponse map[string]AppDetailsEntry

type AppDetailsEntry struct {
	Success bool       `json:"success"`
	Data    AppDetails `json:"data"`
}

type AppDetails struct {
	Name                string     `json:"name"`
	ShortDescription    string     `json:"short_description"`
	DetailedDescription string     `json:"detailed_description"`
	AboutTheGame        string     `json:"about_the_game"`
	HeaderImage         string     `json:"header_image"`
	Website             string     `json:"website"`
	Categories          []Category `json:"categories"`
	Genres              []Genre    `json:"genres"`
}

type Category struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

type Genre struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type GameDetails struct {
	Game
	Genres     string `json:"genres"`     // Raw JSON string from DB
	Categories string `json:"categories"` // Raw JSON string from DB
}
