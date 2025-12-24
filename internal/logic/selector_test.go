package logic

import (
	"testing"

	"github.com/dajoen/steam-pick/internal/model"
)

func TestFilterUnplayed(t *testing.T) {
	games := []model.Game{
		{Name: "Played", PlaytimeForever: 10},
		{Name: "Unplayed", PlaytimeForever: 0},
	}

	unplayed := FilterUnplayed(games)
	if len(unplayed) != 1 {
		t.Errorf("got %d games, want 1", len(unplayed))
	}
	if unplayed[0].Name != "Unplayed" {
		t.Errorf("got %s, want Unplayed", unplayed[0].Name)
	}
}

func TestPickGame(t *testing.T) {
	games := []model.Game{
		{Name: "Game 1"},
		{Name: "Game 2"},
		{Name: "Game 3"},
	}

	// Test deterministic pick
	seed := int64(12345)
	pick1 := PickGame(games, seed)
	pick2 := PickGame(games, seed)

	if pick1.Name != pick2.Name {
		t.Errorf("deterministic pick failed: %s != %s", pick1.Name, pick2.Name)
	}

	// Test empty
	if PickGame(nil, 0) != nil {
		t.Errorf("pick from empty should be nil")
	}
}
