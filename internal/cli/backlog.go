package cli

import (
	"fmt"
	"os"

	"github.com/dajoen/steam-pick/internal/backlog"
	"github.com/spf13/cobra"
)

var (
	failOn string
	output string
	dryRun bool
)

var backlogCmd = &cobra.Command{
	Use:   "backlog",
	Short: "Manage backlog and changelog",
}

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint Backlog.md and CHANGELOG.md",
	Run: func(cmd *cobra.Command, args []string) {
		backlogPath, err := resolveBacklogPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		res, err := backlog.Lint(backlogPath, "CHANGELOG.md")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		hasErrors := len(res.Errors) > 0
		hasWarnings := len(res.Warnings) > 0

		if output == "json" {
			// TODO: Implement JSON output
			fmt.Println("JSON output not implemented yet")
		} else {
			for _, w := range res.Warnings {
				fmt.Printf("WARN: %s\n", w)
			}
			for _, e := range res.Errors {
				fmt.Printf("ERROR: %s\n", e)
			}
		}

		if hasErrors {
			os.Exit(3)
		}
		if hasWarnings && failOn == "warn" {
			os.Exit(2)
		}
		if !hasErrors && !hasWarnings {
			fmt.Println("Backlog is clean.")
		}
	},
}

var backlogSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync completed backlog items to CHANGELOG.md",
	Run: func(cmd *cobra.Command, args []string) {
		backlogPath, err := resolveBacklogPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := backlog.Sync(backlogPath, "CHANGELOG.md", dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(backlogCmd)
	backlogCmd.AddCommand(lintCmd)
	backlogCmd.AddCommand(backlogSyncCmd)

	lintCmd.Flags().StringVar(&failOn, "fail-on", "error", "Fail on 'warn' or 'error'")
	lintCmd.Flags().StringVar(&output, "output", "text", "Output format 'text' or 'json'")

	backlogSyncCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print changes without writing")
}

func resolveBacklogPath() (string, error) {
	candidates := []string{"Backlog.md", "BACKLOG.md"}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	return "", fmt.Errorf("backlog file not found (Backlog.md or BACKLOG.md)")
}
