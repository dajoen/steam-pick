package cli

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/dajoen/steam-pick/internal/db"
	"github.com/spf13/cobra"
)

var (
	profileRecencyHalfLifeDays int
	profileMinHours            int
	profileOutput              string
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Analyze taste profile",
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.New("steam-pick")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = database.Close() }()

		games, err := database.GetGamesWithDetails()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching games: %v\n", err)
			os.Exit(1)
		}

		genreScores := make(map[string]float64)

		now := time.Now().Unix()
		halfLifeSecs := float64(profileRecencyHalfLifeDays * 24 * 3600)

		for _, g := range games {
			playtimeHours := float64(g.PlaytimeForever) / 60.0
			if playtimeHours < float64(profileMinHours) {
				continue
			}

			weight := playtimeHours
			if g.RTimeLastPlayed > 0 && profileRecencyHalfLifeDays > 0 {
				ageSecs := float64(now - int64(g.RTimeLastPlayed))
				decay := math.Pow(0.5, ageSecs/halfLifeSecs)
				weight *= decay
			}

			var genres []struct {
				Description string `json:"description"`
			}
			if err := json.Unmarshal([]byte(g.Genres), &genres); err != nil {
				continue
			}

			for _, genre := range genres {
				genreScores[genre.Description] += weight
			}
		}

		// Sort by score
		type kv struct {
			Key   string
			Value float64
		}
		var ss []kv
		for k, v := range genreScores {
			ss = append(ss, kv{k, v})
		}
		sort.Slice(ss, func(i, j int) bool {
			return ss[i].Value > ss[j].Value
		})

		if profileOutput == "json" {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(ss); err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Taste Profile (Weighted by Playtime & Recency):")
			fmt.Printf("%-30s %s\n", "Genre", "Score")
			fmt.Println("---------------------------------------------")
			for _, kv := range ss {
				fmt.Printf("%-30s %.2f\n", kv.Key, kv.Value)
			}
		}

		profileJSON, _ := json.Marshal(ss)
		if err := database.UpsertTasteProfile("genres", string(profileJSON)); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save profile: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.Flags().IntVar(&profileRecencyHalfLifeDays, "recency-half-life-days", 365, "Half-life in days for recency decay")
	profileCmd.Flags().IntVar(&profileMinHours, "min-hours", 2, "Minimum playtime in hours to consider")
	profileCmd.Flags().StringVar(&profileOutput, "output", "table", "Output format 'table' or 'json'")
}
