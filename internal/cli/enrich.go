package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dajoen/steam-pick/internal/db"
	"github.com/dajoen/steam-pick/internal/model"
	"github.com/dajoen/steam-pick/internal/pcgw"
	"github.com/dajoen/steam-pick/internal/steamapi"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

var (
	enrichRateLimit int
	enrichWorkers   int
	enrichRefresh   bool
)

var enrichCmd = &cobra.Command{
	Use:   "enrich",
	Short: "Enrich game data with store details",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey, err := getAPIKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		client, err := steamapi.NewClient(apiKey, 24*time.Hour, 30*time.Second)
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

		var gamesToEnrich []model.Game
		if enrichRefresh {
			fmt.Println("Refresh enabled: Fetching all owned games...")
			games, err := database.GetOwnedGames()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching games from DB: %v\n", err)
				os.Exit(1)
			}
			gamesToEnrich = games
		} else {
			fmt.Println("Fetching games missing details...")
			games, err := database.GetGamesMissingDetails()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching missing games from DB: %v\n", err)
				os.Exit(1)
			}
			gamesToEnrich = games
		}

		if len(gamesToEnrich) == 0 {
			fmt.Println("No games to enrich.")
			return
		}

		fmt.Printf("Found %d games to enrich.\n", len(gamesToEnrich))

		// rate.Limit is events per second.
		r := rate.Limit(float64(enrichRateLimit) / 60.0)
		limiter := rate.NewLimiter(r, 1)

		// PCGW client
		pcgwClient := pcgw.NewClient()

		var wg sync.WaitGroup
		sem := make(chan struct{}, enrichWorkers)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for i, game := range gamesToEnrich {
			select {
			case <-ctx.Done():
				break
			default:
			}

			wg.Add(1)
			sem <- struct{}{}
			go func(idx int, g model.Game) {
				defer wg.Done()
				defer func() { <-sem }()

				if ctx.Err() != nil {
					return
				}

				if err := limiter.Wait(ctx); err != nil {
					// Context cancelled or limiter error
					return
				}

				fmt.Printf("[%d/%d] Fetching details for %s (%d)...\n", idx+1, len(gamesToEnrich), g.Name, g.AppID)
				details, err := client.GetAppDetails(ctx, g.AppID)
				if err != nil {
					if errors.Is(err, steamapi.ErrRateLimitExceeded) {
						fmt.Fprintf(os.Stderr, "Rate limit exceeded! Stopping enrichment.\n")
						cancel()
						return
					}

					// Try PCGamingWiki fallback
					fmt.Printf("Steam Store failed for %s (%d). Trying PCGamingWiki...\n", g.Name, g.AppID)
					pcgwDetails, pcgwErr := pcgwClient.GetAppDetails(ctx, g.AppID)
					if pcgwErr == nil {
						// Use the name from our DB since PCGW might not return it cleanly
						entry := (*pcgwDetails)[fmt.Sprintf("%d", g.AppID)]
						entry.Data.Name = g.Name
						(*pcgwDetails)[fmt.Sprintf("%d", g.AppID)] = entry

						details = pcgwDetails
						fmt.Printf("Found details for %s on PCGamingWiki.\n", g.Name)
					} else {
						// Don't print error if context was cancelled
						if ctx.Err() == nil {
							fmt.Fprintf(os.Stderr, "Failed to fetch details for %s (%d): %v (PCGW: %v)\n", g.Name, g.AppID, err, pcgwErr)
						}
						// If both fail, we still want to save a stub so we don't retry forever.
						// We create a fake "failed" response which UpsertAppDetails handles by creating a stub.
						details = &model.AppDetailsResponse{
							fmt.Sprintf("%d", g.AppID): {Success: false},
						}
					}
				}

				if err := database.UpsertAppDetails(g.AppID, *details); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to save details for %s (%d): %v\n", g.Name, g.AppID, err)
				}
			}(i, game)
		}

		wg.Wait()
		if ctx.Err() != nil {
			fmt.Println("Enrichment stopped due to errors.")
			os.Exit(1)
		}
		fmt.Println("Enrichment complete.")
	},
}

func init() {
	rootCmd.AddCommand(enrichCmd)
	enrichCmd.Flags().IntVar(&enrichRateLimit, "rate-limit-per-minute", 30, "Rate limit per minute")
	enrichCmd.Flags().IntVar(&enrichWorkers, "workers", 1, "Number of concurrent workers")
	enrichCmd.Flags().BoolVar(&enrichRefresh, "refresh", false, "Refresh existing data")
}
