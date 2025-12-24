package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dajoen/steam-pick/internal/model"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func New(appName string) (*DB, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user cache dir: %w", err)
	}
	dir := filepath.Join(userCacheDir, appName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache dir: %w", err)
	}

	dbPath := filepath.Join(dir, "steampick.db")
	return NewWithDSN(dbPath)
}

func NewWithDSN(dsn string) (*DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	d := &DB{db}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate db: %w", err)
	}

	return d, nil
}

func (d *DB) UpsertGames(games []model.Game) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`
		INSERT INTO owned_games (
			appid, name, playtime_forever, rtime_last_played, img_icon_url,
			has_community_visible_stats, playtime_windows_forever,
			playtime_mac_forever, playtime_linux_forever, playtime_deck_forever,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(appid) DO UPDATE SET
			name=excluded.name,
			playtime_forever=excluded.playtime_forever,
			rtime_last_played=excluded.rtime_last_played,
			img_icon_url=excluded.img_icon_url,
			has_community_visible_stats=excluded.has_community_visible_stats,
			playtime_windows_forever=excluded.playtime_windows_forever,
			playtime_mac_forever=excluded.playtime_mac_forever,
			playtime_linux_forever=excluded.playtime_linux_forever,
			playtime_deck_forever=excluded.playtime_deck_forever,
			updated_at=CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, g := range games {
		_, err := stmt.Exec(
			g.AppID, g.Name, g.PlaytimeForever, g.RTimeLastPlayed, g.ImgIconURL,
			g.HasCommunityVisibleStats, g.PlaytimeWindowsForever,
			g.PlaytimeMacForever, g.PlaytimeLinuxForever, g.PlaytimeDeckForever,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *DB) GetOwnedGames() ([]model.Game, error) {
	rows, err := d.Query("SELECT appid, name, playtime_forever, rtime_last_played FROM owned_games")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []model.Game
	for rows.Next() {
		var g model.Game
		if err := rows.Scan(&g.AppID, &g.Name, &g.PlaytimeForever, &g.RTimeLastPlayed); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, nil
}

func (d *DB) GetGamesMissingDetails() ([]model.Game, error) {
	rows, err := d.Query(`
		SELECT g.appid, g.name, g.playtime_forever, g.rtime_last_played
		FROM owned_games g
		LEFT JOIN app_details ad ON g.appid = ad.appid
		WHERE ad.appid IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []model.Game
	for rows.Next() {
		var g model.Game
		if err := rows.Scan(&g.AppID, &g.Name, &g.PlaytimeForever, &g.RTimeLastPlayed); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, nil
}

func (d *DB) GetGamesWithDetails() ([]model.GameDetails, error) {
	rows, err := d.Query(`
		SELECT
			g.appid, g.name, g.playtime_forever, g.rtime_last_played,
			COALESCE(ad.genres, '[]'), COALESCE(ad.categories, '[]')
		FROM owned_games g
		LEFT JOIN app_details ad ON g.appid = ad.appid
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []model.GameDetails
	for rows.Next() {
		var g model.GameDetails
		if err := rows.Scan(&g.AppID, &g.Name, &g.PlaytimeForever, &g.RTimeLastPlayed, &g.Genres, &g.Categories); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, nil
}

func (d *DB) UpsertAppDetails(appID int, details model.AppDetailsResponse) error {
	key := fmt.Sprintf("%d", appID)
	data, ok := details[key]

	var name, shortDesc, detailedDesc, about, header, website, categories, genres string

	toJSON := func(v interface{}) string {
		b, _ := json.Marshal(v)
		return string(b)
	}

	if ok && data.Success {
		dDetails := data.Data
		name = dDetails.Name
		shortDesc = dDetails.ShortDescription
		detailedDesc = dDetails.DetailedDescription
		about = dDetails.AboutTheGame
		header = dDetails.HeaderImage
		website = dDetails.Website
		categories = toJSON(dDetails.Categories)
		genres = toJSON(dDetails.Genres)
	} else {
		// If success is false, it means the game is delisted or unavailable.
		// We insert a stub record so we don't keep trying to fetch it.
		name = "Unavailable"
		categories = "[]"
		genres = "[]"
	}

	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Exec(`
		INSERT INTO app_details (
			appid, name, short_description, detailed_description, about_the_game,
			header_image, website, categories, genres, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(appid) DO UPDATE SET
			name=excluded.name,
			short_description=excluded.short_description,
			detailed_description=excluded.detailed_description,
			about_the_game=excluded.about_the_game,
			header_image=excluded.header_image,
			website=excluded.website,
			categories=excluded.categories,
			genres=excluded.genres,
			updated_at=CURRENT_TIMESTAMP
	`,
		appID, name, shortDesc, detailedDesc, about,
		header, website, categories, genres,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) UpsertTasteProfile(key string, value string) error {
	_, err := d.Exec(`
		INSERT INTO taste_profile (key, value, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET
			value=excluded.value,
			updated_at=CURRENT_TIMESTAMP
	`, key, value)
	return err
}

func (d *DB) GetTasteProfile(key string) (string, error) {
	var value string
	err := d.QueryRow("SELECT value FROM taste_profile WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (d *DB) GetAppDescription(appID int) (string, error) {
	var desc string
	err := d.QueryRow("SELECT short_description FROM app_details WHERE appid = ?", appID).Scan(&desc)
	if err != nil {
		return "", err
	}
	return desc, nil
}
