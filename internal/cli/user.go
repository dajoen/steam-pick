package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/dajoen/steam-pick/internal/cache"
	"github.com/spf13/viper"
)

type UserCache struct {
	SteamID64 string `json:"steamid64"`
	Vanity    string `json:"vanity"`
}

func getSteamID(ctx context.Context, client SteamClient, steamIDFlag, vanityFlag string) (string, error) {
	// 1. Prefer explicit flags
	if steamIDFlag != "" {
		_ = saveUserCache(steamIDFlag, "")
		return steamIDFlag, nil
	}
	if vanityFlag != "" {
		id, err := client.ResolveVanityURL(ctx, vanityFlag)
		if err != nil {
			return "", err
		}
		_ = saveUserCache(id, vanityFlag)
		return id, nil
	}

	// 2. Check Cache
	c, err := cache.New[UserCache]("steam-pick")
	if err == nil {
		gpgKey := viper.GetString("gpg_key")
		if gpgKey != "" {
			c.WithEncryption(gpgKey)
		}
		// Use a long TTL for user preference (e.g. 30 days)
		if cached, found, _ := c.Get("last_user", 720*time.Hour); found {
			return cached.SteamID64, nil
		}
	}

	// 3. Check Config
	if id := viper.GetString("steamid64"); id != "" {
		return id, nil
	}

	return "", fmt.Errorf("--steamid64 or --vanity is required (or run 'steam-pick login')")
}

func saveUserCache(steamID, vanity string) error {
	c, err := cache.New[UserCache]("steam-pick")
	if err != nil {
		return err
	}
	gpgKey := viper.GetString("gpg_key")
	if gpgKey != "" {
		c.WithEncryption(gpgKey)
	}
	return c.Set("last_user", UserCache{SteamID64: steamID, Vanity: vanity})
}
