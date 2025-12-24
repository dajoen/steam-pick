package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/dajoen/steam-pick/internal/logic"
	"github.com/dajoen/steam-pick/internal/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pickCmd = &cobra.Command{
	Use:   "pick",
	Short: "Pick a random unplayed game",
	Run:   runPick,
}

func init() {
	rootCmd.AddCommand(pickCmd)

	pickCmd.Flags().String("steamid64", "", "SteamID64")
	pickCmd.Flags().String("vanity", "", "Steam Vanity URL name")
	pickCmd.Flags().Bool("include-free-games", false, "Include free games")
	pickCmd.Flags().Int64("seed", 0, "Random seed")
	pickCmd.Flags().Bool("turn-based-only", false, "Only pick turn-based games")
	pickCmd.Flags().Int("max-store-lookups", 200, "Max store lookups for turn-based check")
	pickCmd.Flags().String("country-code", "NL", "Country code for store API")
	pickCmd.Flags().Duration("sleep", 100*time.Millisecond, "Sleep between store calls")
	pickCmd.Flags().Bool("json", false, "Output JSON")
	pickCmd.Flags().Duration("cache-ttl", 24*time.Hour, "Cache TTL")
	pickCmd.Flags().Duration("timeout", 15*time.Second, "HTTP Timeout")
}

func runPick(cmd *cobra.Command, args []string) {
	apiKey, err := getAPIKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	steamID := viper.GetString("steamid64")
	vanity := viper.GetString("vanity")

	ttl, _ := cmd.Flags().GetDuration("cache-ttl")
	timeout, _ := cmd.Flags().GetDuration("timeout")
	includeFree, _ := cmd.Flags().GetBool("include-free-games")
	seed, _ := cmd.Flags().GetInt64("seed")
	turnBased, _ := cmd.Flags().GetBool("turn-based-only")
	maxLookups, _ := cmd.Flags().GetInt("max-store-lookups")
	country, _ := cmd.Flags().GetString("country-code")
	sleep, _ := cmd.Flags().GetDuration("sleep")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	vanityTTL := viper.GetDuration("auth_cache_ttl")

	client, err := NewSteamClient(apiKey, ttl, vanityTTL, timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	steamID, err = getSteamID(ctx, client, steamID, vanity)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	games, err := client.GetOwnedGames(ctx, steamID, includeFree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching games: %v\n", err)
		os.Exit(1)
	}

	unplayed := logic.FilterUnplayed(games)
	if len(unplayed) == 0 {
		fmt.Fprintln(os.Stderr, "No unplayed games found.")
		os.Exit(0)
	}

	var picked *model.Game

	if turnBased {
		storeClient := NewStoreClient(timeout)

		// Shuffle unplayed list to check random games
		shuffled := make([]model.Game, len(unplayed))
		copy(shuffled, unplayed)

		var src rand.Source
		if seed != 0 {
			src = rand.NewSource(seed)
		} else {
			src = rand.NewSource(time.Now().UnixNano())
		}
		r := rand.New(src)

		r.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		lookups := 0
		for i := range shuffled {
			if lookups >= maxLookups {
				break
			}

			isTurn, err := storeClient.IsTurnBased(ctx, shuffled[i].AppID, country)
			if err == nil && isTurn {
				picked = &shuffled[i]
				picked.IsTurnBased = true
				break
			}

			lookups++
			time.Sleep(sleep)
		}

		if picked == nil {
			fmt.Fprintln(os.Stderr, "No turn-based game found within limit, falling back to random unplayed game.")
		}
	}

	if picked == nil {
		picked = logic.PickGame(unplayed, seed)
	}

	if picked == nil {
		// Should not happen if unplayed > 0
		fmt.Fprintln(os.Stderr, "Failed to pick a game.")
		os.Exit(1)
	}

	picked.StoreURL = fmt.Sprintf("https://store.steampowered.com/app/%d", picked.AppID)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(picked); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Name: %s\n", picked.Name)
		fmt.Printf("AppID: %d\n", picked.AppID)
		fmt.Printf("Store URL: %s\n", picked.StoreURL)
		if picked.IsTurnBased {
			fmt.Println("Turn-based: Yes")
		}
	}
}
