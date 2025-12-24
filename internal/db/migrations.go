package db

import "fmt"

type migration struct {
	version int
	up      string
}

var migrations = []migration{
	{
		version: 1,
		up: `
		CREATE TABLE IF NOT EXISTS owned_games (
			appid INTEGER PRIMARY KEY,
			name TEXT,
			playtime_forever INTEGER,
			rtime_last_played INTEGER,
			img_icon_url TEXT,
			has_community_visible_stats BOOLEAN,
			playtime_windows_forever INTEGER,
			playtime_mac_forever INTEGER,
			playtime_linux_forever INTEGER,
			playtime_deck_forever INTEGER,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS app_details (
			appid INTEGER PRIMARY KEY,
			name TEXT,
			short_description TEXT,
			detailed_description TEXT,
			about_the_game TEXT,
			header_image TEXT,
			website TEXT,
			pc_requirements TEXT, -- JSON
			mac_requirements TEXT, -- JSON
			linux_requirements TEXT, -- JSON
			developers TEXT, -- JSON array
			publishers TEXT, -- JSON array
			price_overview TEXT, -- JSON
			platforms TEXT, -- JSON
			categories TEXT, -- JSON array
			genres TEXT, -- JSON array
			screenshots TEXT, -- JSON array
			movies TEXT, -- JSON array
			recommendations TEXT, -- JSON
			achievements TEXT, -- JSON
			release_date TEXT, -- JSON
			support_info TEXT, -- JSON
			background TEXT,
			content_descriptors TEXT, -- JSON
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS taste_profile (
			key TEXT PRIMARY KEY,
			value TEXT, -- JSON
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS recommendations (
			appid INTEGER PRIMARY KEY,
			score REAL,
			reason TEXT,
			type TEXT, -- 'backlog' or 'discovery'
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`,
	},
}

func (d *DB) migrate() error {
	// Bootstrap schema_migrations if not exists
	_, err := d.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}

	var currentVersion int
	err = d.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&currentVersion)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if m.version > currentVersion {
			fmt.Printf("Applying migration %d...\n", m.version)
			if _, err := d.Exec(m.up); err != nil {
				return fmt.Errorf("migration %d failed: %w", m.version, err)
			}
			if _, err := d.Exec("INSERT INTO schema_migrations (version) VALUES (?)", m.version); err != nil {
				return fmt.Errorf("failed to record migration %d: %w", m.version, err)
			}
		}
	}

	return nil
}
