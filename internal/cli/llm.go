package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/dajoen/steam-pick/internal/llm"
	"github.com/spf13/cobra"
)

var (
	llmBaseURL    string
	llmModel      string
	llmEmbedModel string
)

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Manage LLM integration",
}

var llmCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check LLM connection",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := llm.Config{
			BaseURL:    llmBaseURL,
			Model:      llmModel,
			EmbedModel: llmEmbedModel,
		}
		client := llm.NewOllamaClient(cfg)

		fmt.Printf("Checking connection to %s...\n", llmBaseURL)
		if err := client.Check(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Connection successful.")

		if llmModel != "" {
			fmt.Printf("Checking generation with model %s...\n", llmModel)
			res, err := client.Generate(context.Background(), "Hello")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Response: %s\n", res)
		}
	},
}

func init() {
	rootCmd.AddCommand(llmCmd)
	llmCmd.AddCommand(llmCheckCmd)

	llmCheckCmd.Flags().StringVar(&llmBaseURL, "base-url", "http://localhost:11434", "LLM Base URL")
	llmCheckCmd.Flags().StringVar(&llmModel, "model", "", "Model name for generation check")
	llmCheckCmd.Flags().StringVar(&llmEmbedModel, "embed-model", "", "Model name for embedding check")
}
