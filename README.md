# steam-pick

A production-grade CLI tool that recommends an unplayed Steam game (0 minutes) from your Steam library using the Steam Web API.

## Features

- **List Unplayed Games**: Fetch and filter your library for games with 0 playtime.
- **Pick a Game**: Randomly select a game to play, with optional "turn-based" filtering.
- **Secure Secrets**: Native integration with `gopass` to securely retrieve your Steam API Key.
- **Resilience**: Automatic retries with exponential backoff for transient API failures.
- **Caching**: Caches API responses to avoid hitting rate limits and improve speed.
- **Cross-Platform**: Single binary for Linux, macOS, and Windows.

## Prerequisites

You need a Steam Web API Key. You can get one here: [https://steamcommunity.com/dev/apikey](https://steamcommunity.com/dev/apikey)

## Installation

### From Binary

Download the latest release from the [Releases](https://github.com/dajoen/steam-pick/releases) page.

### From Source

```bash
go install github.com/dajoen/steam-pick/cmd/steam-pick@latest
```

## Usage

### Configuration

The tool looks for configuration in the following order (highest precedence first):
1. Command-line flags (e.g., `--api-key`, `--gopass-path`)
2. Environment variables (e.g., `STEAM_API_KEY`)
3. Config file (`$HOME/.steam-pick.yaml`)

You can provide the API key via:
- **Flag**: `--api-key "..."`
- **Env**: `export STEAM_API_KEY="..."`
- **Gopass**: `--gopass-path "steam/api-key"` (supports interactive pinentry)

```bash
export STEAM_API_KEY="your-api-key"
export STEAM_STEAMID64="your-steam-id" # Optional default
```

### Commands

#### List unplayed games

```bash
steam-pick list --steamid64 <your-steam-id>
steam-pick list --vanity <your-vanity-url-name>
steam-pick list --include-free-games
# Using gopass
steam-pick list --gopass-path steam/api-key --vanity <your-vanity-url-name>
```

#### Pick a random game

```bash
steam-pick pick --steamid64 <your-steam-id>
steam-pick pick --turn-based-only
steam-pick pick --seed 12345 # Deterministic pick
```

#### Manage Cache

```bash
steam-pick cache
steam-pick cache --clear
```

## Troubleshooting

- **Empty List**: Ensure your Steam profile Game Details are set to **Public**.
- **Rate Limiting**: The tool caches responses. If you hit limits, wait a few minutes.
- **Turn-based detection**: This is a heuristic based on store tags. It may not be 100% accurate.

## Development

### Build

```bash
make build
```

### Test

```bash
make test
```

### Lint

```bash
make lint
```
