package cli

import (
	"fmt"
	"os"

	"github.com/dajoen/steam-pick/internal/cache"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage cache",
	Run:   runCache,
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.Flags().Bool("clear", false, "Clear cache")
}

func runCache(cmd *cobra.Command, args []string) {
	// We can use any type for cache init since we just want to manage the dir
	c, err := cache.New[any]("steam-pick")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing cache: %v\n", err)
		os.Exit(1)
	}

	clear, _ := cmd.Flags().GetBool("clear")
	if clear {
		if err := c.Clear(); err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing cache: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Cache cleared.")
		return
	}

	count, size, err := c.Stats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting cache stats: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Cache Directory: %s\n", c.DirPath())
	fmt.Printf("Files: %d\n", count)
	fmt.Printf("Size: %d bytes\n", size)
}
