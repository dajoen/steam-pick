package logic

import (
	"math/rand"
	"time"

	"github.com/jeroenverhoeven/steam-pick/internal/model"
)

// FilterUnplayed returns games with 0 playtime.
func FilterUnplayed(games []model.Game) []model.Game {
	var unplayed []model.Game
	for _, g := range games {
		if g.PlaytimeForever == 0 {
			unplayed = append(unplayed, g)
		}
	}
	return unplayed
}

// PickGame selects a random game from the list.
// If seed is non-zero, it uses it for deterministic selection.
func PickGame(games []model.Game, seed int64) *model.Game {
	if len(games) == 0 {
		return nil
	}

	var src rand.Source
	if seed != 0 {
		src = rand.NewSource(seed)
	} else {
		src = rand.NewSource(time.Now().UnixNano())
	}
	r := rand.New(src)

	idx := r.Intn(len(games))
	return &games[idx]
}
