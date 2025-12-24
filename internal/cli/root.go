package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	apiKey  string
)

var rootCmd = &cobra.Command{
	Use:   "steam-unplayed",
	Short: "Recommend an unplayed Steam game from your library",
	Long: `steam-unplayed is a CLI tool that helps you find games in your Steam library
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.steam-unplayed.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Steam Web API Key")
	
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
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
		viper.SetConfigName(".steam-unplayed")
	}

	viper.SetEnvPrefix("STEAM")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		// Config file found and read
	}
}
