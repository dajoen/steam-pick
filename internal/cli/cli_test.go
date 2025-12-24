package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/dajoen/steam-pick/internal/model"
	"github.com/spf13/viper"
)

type MockSteamClient struct {
	Games []model.Game
}

func (m *MockSteamClient) ResolveVanityURL(ctx context.Context, vanityURL string) (string, error) {
	return "76561198000000000", nil
}

func (m *MockSteamClient) GetOwnedGames(ctx context.Context, steamID64 string, includeFree bool) ([]model.Game, error) {
	return m.Games, nil
}

func TestListCommand(t *testing.T) {
	// Mock dependencies
	oldFactory := NewSteamClient
	defer func() { NewSteamClient = oldFactory }()

	NewSteamClient = func(apiKey string, ttl, timeout time.Duration) (SteamClient, error) {
		return &MockSteamClient{
			Games: []model.Game{
				{AppID: 1, Name: "Unplayed Game", PlaytimeForever: 0},
				{AppID: 2, Name: "Played Game", PlaytimeForever: 100},
			},
		}, nil
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Setup Viper
	viper.Set("api_key", "test-key")
	viper.Set("steamid64", "76561198000000000")

	// Run command
	// We need to reset flags or use a new command instance, but listCmd is global.
	// For this simple test, it's fine.
	listCmd.Run(listCmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if output != "1: Unplayed Game\n" {
		t.Errorf("Expected output '1: Unplayed Game\\n', got '%s'", output)
	}
}
