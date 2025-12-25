package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/dajoen/steam-pick/internal/db"
	"github.com/dajoen/steam-pick/internal/llm"
	"github.com/spf13/cobra"
)

var (
	recommendMode    string
	recommendTop     int
	recommendExplain bool
	recommendOutput  string
)

var recommendCmd = &cobra.Command{
	Use:   "recommend",
	Short: "Recommend games based on taste profile",
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.New("steam-pick")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer database.Close()

		// Load profile
		profileJSON, err := database.GetTasteProfile("genres")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading profile (run 'profile' first): %v\n", err)
			os.Exit(1)
		}

		type GenreScore struct {
			Key   string
			Value float64
		}
		var profile []GenreScore
		if err := json.Unmarshal([]byte(profileJSON), &profile); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing profile: %v\n", err)
			os.Exit(1)
		}

		profileMap := make(map[string]float64)
		for _, p := range profile {
			profileMap[p.Key] = p.Value
		}

		// Load candidates
		games, err := database.GetGamesWithDetails()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching games: %v\n", err)
			os.Exit(1)
		}

		type Recommendation struct {
			AppID       int     `json:"appid"`
			Name        string  `json:"name"`
			Score       float64 `json:"score"`
			Explanation string  `json:"explanation,omitempty"`
		}

		var recommendations []Recommendation

		for _, g := range games {
			// Backlog mode: playtime < 2 hours (120 mins)
			if recommendMode == "backlog" && g.PlaytimeForever > 120 {
				continue
			}

			var genres []struct {
				Description string `json:"description"`
			}
			_ = json.Unmarshal([]byte(g.Genres), &genres)

			score := 0.0
			for _, genre := range genres {
				if s, ok := profileMap[genre.Description]; ok {
					score += s
				}
			}

			if score > 0 {
				recommendations = append(recommendations, Recommendation{
					AppID: g.AppID,
					Name:  g.Name,
					Score: score,
				})
			}
		}

		// Sort
		sort.Slice(recommendations, func(i, j int) bool {
			return recommendations[i].Score > recommendations[j].Score
		})

		// Top N
		if recommendTop < 0 {
			fmt.Fprintln(os.Stderr, "Error: --top must be >= 0")
			os.Exit(1)
		}
		if len(recommendations) > recommendTop {
			recommendations = recommendations[:recommendTop]
		}

		// Explain
		if recommendExplain {
			cfg := llm.Config{
				BaseURL:    llmBaseURL,
				Model:      llmModel,
				EmbedModel: llmEmbedModel,
			}
			if cfg.BaseURL == "" {
				cfg.BaseURL = "http://localhost:11434"
			}

			client := llm.NewOllamaClient(cfg)

			// Get top 5 genres from profile for context
			var topGenres []string
			for i := 0; i < 5 && i < len(profile); i++ {
				topGenres = append(topGenres, profile[i].Key)
			}

			for i := range recommendations {
				rec := &recommendations[i]
				desc, _ := database.GetAppDescription(rec.AppID)

				prompt := fmt.Sprintf(
					"I like %s. Why should I play %s? It is described as: %s. Keep it short.",
					strings.Join(topGenres, ", "),
					rec.Name,
					desc,
				)

				fmt.Printf("Generating explanation for %s...\n", rec.Name)
				expl, err := client.Generate(context.Background(), prompt)
				if err == nil {
					rec.Explanation = strings.TrimSpace(expl)
				} else {
					fmt.Fprintf(os.Stderr, "LLM error: %v\n", err)
				}
			}
		}

		if recommendOutput == "json" {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(recommendations); err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Recommendations:")
			fmt.Printf("%-10s %-40s %s\n", "AppID", "Name", "Score")
			fmt.Println("------------------------------------------------------------")
			for _, r := range recommendations {
				fmt.Printf("%-10d %-40s %.2f\n", r.AppID, r.Name, r.Score)
				if r.Explanation != "" {
					fmt.Printf("  Explanation: %s\n", r.Explanation)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(recommendCmd)
	recommendCmd.Flags().StringVar(&recommendMode, "mode", "backlog", "Mode: 'backlog' or 'discovery'")
	recommendCmd.Flags().IntVar(&recommendTop, "top", 10, "Number of recommendations")
	recommendCmd.Flags().BoolVar(&recommendExplain, "explain", false, "Explain recommendations using LLM")
	recommendCmd.Flags().StringVar(&recommendOutput, "output", "table", "Output format 'table' or 'json'")

	recommendCmd.Flags().StringVar(&llmBaseURL, "llm-base-url", "http://localhost:11434", "LLM Base URL")
	recommendCmd.Flags().StringVar(&llmModel, "llm-model", "llama3", "LLM Model")
}
