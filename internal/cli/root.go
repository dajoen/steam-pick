package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	apiKey     string
	gopassPath string
)

var rootCmd = &cobra.Command{
	Use:   "steam-pick",
	Short: "Recommend an unplayed Steam game from your library",
	Long: `steam-pick is a CLI tool that helps you find games in your Steam library
that you haven't played yet (0 minutes playtime).`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.steam-pick.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Steam Web API Key")
	rootCmd.PersistentFlags().StringVar(&gopassPath, "gopass-path", "", "Gopass path to Steam API Key (e.g. steam/api-key)")
	
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("gopass_path", rootCmd.PersistentFlags().Lookup("gopass-path"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".steam-pick")
	}

	viper.SetEnvPrefix("STEAM")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		// Config file found and read
	}
}

func getAPIKey() (string, error) {
	key := viper.GetString("api_key")
	if key != "" {
		return key, nil
	}

	path := viper.GetString("gopass_path")
	if path != "" {
		// Check if gopass is installed
		if _, err := exec.LookPath("gopass"); err != nil {
			return "", fmt.Errorf("gopass not found in PATH")
		}

		cmd := exec.Command("gopass", "show", "-o", path)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get key from gopass: %w", err)
		}
		return strings.TrimSpace(string(out)), nil
	}

	return "", fmt.Errorf("api key not found (use --api-key, STEAM_API_KEY, or --gopass-path)")
}
