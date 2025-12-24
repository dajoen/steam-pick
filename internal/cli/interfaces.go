package cli

import (
	"context"
	"time"

	"github.com/dajoen/steam-pick/internal/model"
	"github.com/dajoen/steam-pick/internal/steamapi"
	"github.com/dajoen/steam-pick/internal/storeapi"
)

// SteamClient defines the interface for Steam Web API interactions.
type SteamClient interface {
	ResolveVanityURL(ctx context.Context, vanityURL string) (string, error)
	GetOwnedGames(ctx context.Context, steamID64 string, includeFree bool) ([]model.Game, error)
}

// StoreClient defines the interface for Steam Store API interactions.
type StoreClient interface {
	IsTurnBased(ctx context.Context, appID int, country string) (bool, error)
}

// ClientFactoryFunc is a function that creates a SteamClient.
type ClientFactoryFunc func(apiKey string, ttl, timeout time.Duration) (SteamClient, error)

// StoreClientFactoryFunc is a function that creates a StoreClient.
type StoreClientFactoryFunc func(timeout time.Duration) StoreClient

// Default factories
var (
	NewSteamClient ClientFactoryFunc = func(apiKey string, ttl, timeout time.Duration) (SteamClient, error) {
		return steamapi.NewClient(apiKey, ttl, timeout)
	}
	NewStoreClient StoreClientFactoryFunc = func(timeout time.Duration) StoreClient {
		return storeapi.NewClient(timeout)
	}
)
