# steam-unplayed

A production-grade CLI tool that recommends an unplayed Steam game (0 minutes) from your Steam library using the Steam Web API.

## Features

- **List Unplayed Games**: Fetch and filter your library for games with 0 playtime.
- **Pick a Game**: Randomly select a game to play, with optional "turn-based" filtering.
- **Caching**: Caches API responses to avoid hitting rate limits and improve speed.
- **Cross-Platform**: Single binary for Linux, macOS, and Windows.

## Prerequisites

You need a Steam Web API Key. You can get one here: [https://steamcommunity.com/dev/apikey](https://steamcommunity.com/dev/apikey)

## Installation

### From Binary

Download the latest release from the [Releases](https://github.com/jeroenverhoeven/steam-pick/releases) page.

### From Source

```bash
go install github.com/jeroenverhoeven/steam-pick/cmd/steam-unplayed@latest
```

## Usage

### Configuration

You can provide the API key via:
- Flag `--api-key`
- Environment variable `STEAM_API_KEY`
- Gopass path via `--gopass-path` (e.g. `steam/api-key`)

```bash
export STEAM_API_KEY="your-api-key"
export STEAM_STEAMID64="your-steam-id" # Optional default
```

### Commands

#### List unplayed games

```bash
steam-unplayed list --steamid64 <your-steam-id>
steam-unplayed list --vanity <your-vanity-url-name>
steam-unplayed list --include-free-games
# Using gopass
steam-unplayed list --gopass-path steam/api-key --vanity <your-vanity-url-name>
```

#### Pick a random game

```bash
steam-unplayed pick --steamid64 <your-steam-id>
steam-unplayed pick --turn-based-only
steam-unplayed pick --seed 12345 # Deterministic pick
```

#### Manage Cache

```bash
steam-unplayed cache
steam-unplayed cache --clear
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
