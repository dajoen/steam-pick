package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dajoen/steam-pick/internal/logic"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List unplayed games",
	Run:   runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().String("steamid64", "", "SteamID64")
	listCmd.Flags().String("vanity", "", "Steam Vanity URL name")
	listCmd.Flags().Bool("include-free-games", false, "Include free games")
	listCmd.Flags().Int("limit", 50, "Limit output")
	listCmd.Flags().Bool("json", false, "Output JSON")
	listCmd.Flags().Duration("cache-ttl", 24*time.Hour, "Cache TTL")
	listCmd.Flags().Duration("timeout", 15*time.Second, "HTTP Timeout")

	_ = viper.BindPFlag("steamid64", listCmd.Flags().Lookup("steamid64"))
	_ = viper.BindPFlag("vanity", listCmd.Flags().Lookup("vanity"))
}

func runList(cmd *cobra.Command, args []string) {
	apiKey, err := getAPIKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	steamID := viper.GetString("steamid64")
	vanity := viper.GetString("vanity")

	ttl, _ := cmd.Flags().GetDuration("cache-ttl")
	timeout, _ := cmd.Flags().GetDuration("timeout")
	vanityTTL := viper.GetDuration("auth_cache_ttl")
	includeFree, _ := cmd.Flags().GetBool("include-free-games")
	limit, _ := cmd.Flags().GetInt("limit")
	jsonOutput, _ := cmd.Flags().GetBool("json")

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
		if jsonOutput {
			fmt.Println("[]")
		} else {
			fmt.Println("No unplayed games found (or profile is private).")
		}
		return
	}

	if limit > 0 && len(unplayed) > limit {
		unplayed = unplayed[:limit]
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(unplayed); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		for _, g := range unplayed {
			fmt.Printf("%d: %s\n", g.AppID, g.Name)
		}
	}
}
