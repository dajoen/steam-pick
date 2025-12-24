package db_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/dajoen/steam-pick/internal/db"
	"github.com/dajoen/steam-pick/internal/model"
	"github.com/mattn/go-sqlite3"
)

func TestNewWithDSN(t *testing.T) {
	// Test creating a new DB with an in-memory DSN
	d, err := db.NewWithDSN(":memory:")
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer d.Close()

	// Verify that the tables were created
	tables := []string{"owned_games", "app_details", "taste_profile"}
	for _, table := range tables {
		var name string
		err := d.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("Table %s not found: %v", table, err)
		}
	}
}

func TestUpsertGames(t *testing.T) {
	d, err := db.NewWithDSN(":memory:")
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer d.Close()

	games := []model.Game{
		{
			AppID:           10,
			Name:            "Counter-Strike",
			PlaytimeForever: 1000,
			RTimeLastPlayed: 1600000000,
		},
		{
			AppID:           20,
			Name:            "Team Fortress Classic",
			PlaytimeForever: 500,
			RTimeLastPlayed: 1500000000,
		},
	}

	err = d.UpsertGames(games)
	if err != nil {
		t.Fatalf("UpsertGames failed: %v", err)
	}

	// Verify games were inserted
	storedGames, err := d.GetGamesWithDetails()
	if err != nil {
		t.Fatalf("GetGamesWithDetails failed: %v", err)
	}

	if len(storedGames) != 2 {
		t.Errorf("Expected 2 games, got %d", len(storedGames))
	}

	// Verify details of one game
	found := false
	for _, g := range storedGames {
		if g.AppID == 10 {
			found = true
			if g.Name != "Counter-Strike" {
				t.Errorf("Expected name Counter-Strike, got %s", g.Name)
			}
			if g.PlaytimeForever != 1000 {
				t.Errorf("Expected playtime 1000, got %d", g.PlaytimeForever)
			}
		}
	}
	if !found {
		t.Error("Game with AppID 10 not found")
	}
}

func TestUpsertAppDetails(t *testing.T) {
	d, err := db.NewWithDSN(":memory:")
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer d.Close()

	appID := 10
	details := model.AppDetails{
		Name:                "Counter-Strike",
		ShortDescription:    "Action game",
		DetailedDescription: "Detailed action game",
		AboutTheGame:        "About the game",
		HeaderImage:         "http://example.com/header.jpg",
		Website:             "http://example.com",
		Categories: []model.Category{
			{ID: 1, Description: "Multi-player"},
		},
		Genres: []model.Genre{
			{ID: "1", Description: "Action"},
		},
	}

	response := model.AppDetailsResponse{
		fmt.Sprintf("%d", appID): model.AppDetailsEntry{
			Success: true,
			Data:    details,
		},
	}

	err = d.UpsertAppDetails(appID, response)
	if err != nil {
		t.Fatalf("UpsertAppDetails failed: %v", err)
	}

	// Verify we can retrieve the description
	desc, err := d.GetAppDescription(appID)
	if err != nil {
		t.Fatalf("GetAppDescription failed: %v", err)
	}

	if desc != "Action game" {
		t.Errorf("Expected description 'Action game', got '%s'", desc)
	}
}

func TestUpsertTasteProfile(t *testing.T) {
	d, err := db.NewWithDSN(":memory:")
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer d.Close()

	key := "user_123"
	value := `{"likes": ["Action"], "dislikes": ["Strategy"]}`

	err = d.UpsertTasteProfile(key, value)
	if err != nil {
		t.Fatalf("UpsertTasteProfile failed: %v", err)
	}

	// Verify retrieval
	retrievedValue, err := d.GetTasteProfile(key)
	if err != nil {
		t.Fatalf("GetTasteProfile failed: %v", err)
	}

	if retrievedValue != value {
		t.Errorf("Expected value '%s', got '%s'", value, retrievedValue)
	}

	// Test non-existent key
	_, err = d.GetTasteProfile("non_existent")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestUniqueConstraint(t *testing.T) {
	d, err := db.NewWithDSN(":memory:")
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer d.Close()

	games := []model.Game{
		{AppID: 1, Name: "Game 1"},
	}

	// Insert once
	if err := d.UpsertGames(games); err != nil {
		t.Fatalf("First insert failed: %v", err)
	}

	// Insert again (should update, not fail)
	games[0].Name = "Game 1 Updated"
	if err := d.UpsertGames(games); err != nil {
		t.Fatalf("Second insert failed: %v", err)
	}

	storedGames, _ := d.GetGamesWithDetails()
	if len(storedGames) != 1 {
		t.Errorf("Expected 1 game, got %d", len(storedGames))
	}
	if storedGames[0].Name != "Game 1 Updated" {
		t.Errorf("Expected updated name, got %s", storedGames[0].Name)
	}
}

func TestConcurrentAccess(t *testing.T) {
	// SQLite with :memory: and cache=shared or just default might have locking issues if not handled
	// but go-sqlite3 handles locking. This test ensures we don't panic or deadlock easily.
	d, err := db.NewWithDSN("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer d.Close()

	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			_ = d.UpsertTasteProfile("key", fmt.Sprintf("value-%d", i))
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = d.GetTasteProfile("key")
		}
		done <- true
	}()

	<-done
	<-done
}

func TestSqliteError(t *testing.T) {
	// Test handling of SQLite specific errors if needed
	// For example, trying to open a directory as a DB file
	_, err := db.NewWithDSN("/")
	if err == nil {
		t.Error("Expected error opening directory as DB, got nil")
	}
	// Check if it's a sqlite error
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		t.Logf("Got sqlite error: %v", sqliteErr)
	}
}

func TestGetGamesMissingDetails(t *testing.T) {
	d, err := db.NewWithDSN(":memory:")
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	defer d.Close()

	games := []model.Game{
		{AppID: 1, Name: "Game 1"},
		{AppID: 2, Name: "Game 2"},
	}
	if err := d.UpsertGames(games); err != nil {
		t.Fatalf("UpsertGames failed: %v", err)
	}

	// Insert details for Game 1
	details := model.AppDetailsResponse{
		"1": model.AppDetailsEntry{
			Success: true,
			Data:    model.AppDetails{Name: "Game 1"},
		},
	}
	if err := d.UpsertAppDetails(1, details); err != nil {
		t.Fatalf("UpsertAppDetails failed: %v", err)
	}

	// Should return Game 2
	missing, err := d.GetGamesMissingDetails()
	if err != nil {
		t.Fatalf("GetGamesMissingDetails failed: %v", err)
	}

	if len(missing) != 1 {
		t.Errorf("Expected 1 game missing details, got %d", len(missing))
	}
	if len(missing) > 0 && missing[0].AppID != 2 {
		t.Errorf("Expected Game 2 to be missing details, got Game %d", missing[0].AppID)
	}
}
