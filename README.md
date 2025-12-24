# steam-pick

A production-grade CLI tool that recommends an unplayed Steam game (0 minutes) from your Steam library using the Steam Web API.

## Features

- **List Unplayed Games**: Fetch and filter your library for games with 0 playtime.
- **Pick a Game**: Randomly select a game to play, with optional "turn-based" filtering.
- **Secure Configuration**: Retrieve your Steam API key securely from `gopass`.
- **Resilient**: Built-in retries and caching to handle API flakiness and rate limits.
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

You can provide the API key via:
- Flag `--api-key`
- Environment variable `STEAM_API_KEY`
- Gopass path via `--gopass-path` (e.g. `steam/api-key`)
- Config file `$HOME/.steam-pick.yaml`

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

## Advanced Features (Local LLM & Recommendations)

### 1. Sync Library
Fetch your games into a local SQLite database.
This command compares your library with the database and only updates changed records.
```bash
steam-pick sync --steamid <your-steam-id>
```

### 2. Enrich Data
Fetch store details (genres, descriptions) for your games.
This command is idempotent and will skip games that already have details.
Use `--refresh` to force an update of all games.
```bash
steam-pick enrich --workers 5
steam-pick enrich --refresh # Force update all games
```

### 3. Build Taste Profile
Analyze your playtime to understand your preferences.
```bash
steam-pick profile
```

### 4. Get Recommendations
Get recommendations from your backlog based on your profile.
```bash
steam-pick recommend --top 5 --explain
```

### Local LLM Setup
To use LLM features (explanation), you need a local LLM running (e.g. Ollama).
```bash
ollama run llama3
steam-pick llm check --model llama3
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

### CI Integration

It is recommended to run the backlog linter in CI:
```bash
steam-pick backlog lint --fail-on error
```

### Project Management

This project uses [Backlog.md](https://github.com/MrLesk/Backlog.md) for task management.
The backlog is stored in the `backlog/` directory.

To view the board:
```bash
backlog board
```

To create a task:
```bash
backlog task create "Task Title"
```

## Development

### Pre-commit Hooks

This project uses [pre-commit](https://pre-commit.com/) to ensure code quality.

1. Install pre-commit:
   ```bash
   brew install pre-commit  # macOS
   pip install pre-commit   # Cross-platform
   ```

2. Install the hooks:
   ```bash
   pre-commit install
   ```

3. Run hooks manually (optional):
   ```bash
   pre-commit run --all-files
   ```
