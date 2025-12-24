package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dajoen/steam-pick/internal/db"
	"github.com/dajoen/steam-pick/internal/model"
	"github.com/dajoen/steam-pick/internal/steamapi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	syncSteamID           string
	syncVanity            string
	syncIncludeFreeToPlay bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync Steam library to local database",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey, err := getAPIKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		vanityTTL := viper.GetDuration("auth_cache_ttl")
		client, err := steamapi.NewClient(apiKey, 24*time.Hour, vanityTTL, 30*time.Second)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		syncSteamID, err = getSteamID(context.Background(), client, syncSteamID, syncVanity)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		database, err := db.New("steam-pick")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer database.Close()

		fmt.Printf("Fetching games for SteamID: %s\n", syncSteamID)
		games, err := client.GetOwnedGames(context.Background(), syncSteamID, syncIncludeFreeToPlay)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching games: %v\n", err)
			os.Exit(1)
		}

		// Fetch existing games to compare
		existingGames, err := database.GetOwnedGames()
		if err != nil {
			// If error (e.g. empty table), just proceed
			existingGames = []model.Game{}
		}

		existingMap := make(map[int]model.Game)
		for _, g := range existingGames {
			existingMap[g.AppID] = g
		}

		var newGames, updatedGames []model.Game
		unchangedCount := 0

		for _, g := range games {
			if existing, ok := existingMap[g.AppID]; !ok {
				newGames = append(newGames, g)
			} else {
				// Check if updated
				if g.PlaytimeForever != existing.PlaytimeForever || g.RTimeLastPlayed != existing.RTimeLastPlayed {
					updatedGames = append(updatedGames, g)
				} else {
					unchangedCount++
				}
			}
		}

		fmt.Printf("Sync Summary: %d new, %d updated, %d unchanged.\n", len(newGames), len(updatedGames), unchangedCount)

		gamesToSave := append(newGames, updatedGames...)
		if len(gamesToSave) > 0 {
			fmt.Printf("Saving %d games to database...\n", len(gamesToSave))
			if err := database.UpsertGames(gamesToSave); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving games: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Database is already up to date.")
		}

		fmt.Println("Sync complete.")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringVar(&syncSteamID, "steamid", "", "SteamID64")
	syncCmd.Flags().StringVar(&syncVanity, "vanity", "", "Steam Vanity URL name")
	syncCmd.Flags().BoolVar(&syncIncludeFreeToPlay, "include-free-to-play", false, "Include free-to-play games")
}
