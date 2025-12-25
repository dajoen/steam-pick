package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Steam and save credentials",
	Run:   runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) {
	if err := Login(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
		os.Exit(1)
	}
}

// Login performs the interactive login flow.
func Login(ctx context.Context) error {
	reader := bufio.NewReader(os.Stdin)

	// 1. API Key
	apiKey := viper.GetString("api_key")
	if apiKey == "" {
		fmt.Print("Enter Steam Web API Key: ")
		input, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(input)
	}

	if apiKey == "" {
		return fmt.Errorf("API Key is required")
	}

	// 2. User ID
	steamID := viper.GetString("steamid64")
	vanity := viper.GetString("vanity")

	if steamID == "" && vanity == "" {
		fmt.Print("Enter SteamID64 or Vanity URL name: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if len(input) == 17 && isNumeric(input) {
			steamID = input
		} else {
			vanity = input
		}
	}

	// 3. Validate
	fmt.Println("Validating credentials...")
	client, err := NewSteamClient(apiKey, time.Minute, time.Minute, 10*time.Second)
	if err != nil {
		return err
	}

	resolvedID := steamID
	if vanity != "" {
		id, err := client.ResolveVanityURL(ctx, vanity)
		if err != nil {
			return fmt.Errorf("failed to resolve vanity URL: %w", err)
		}
		resolvedID = id
	}

	// Verify by fetching games (lightweight check)
	_, err = client.GetOwnedGames(ctx, resolvedID, false)
	if err != nil {
		return fmt.Errorf("failed to verify credentials (could not fetch games): %w", err)
	}

	// 4. Save
	viper.Set("api_key", apiKey)
	viper.Set("steamid64", resolvedID)
	if vanity != "" {
		viper.Set("vanity", vanity)
	}

	// Determine config path
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configFile = filepath.Join(home, ".steam-pick.yaml")
	}

	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", configFile, err)
	}

	fmt.Printf("Login successful! Configuration saved to %s\n", configFile)
	return nil
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
