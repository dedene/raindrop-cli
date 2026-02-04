# 🔖 raindrop-cli - Raindrop in your terminal

[![CI](https://github.com/dedene/raindrop-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/dedene/raindrop-cli/actions/workflows/ci.yml)
[![Go 1.23+](https://img.shields.io/badge/go-1.23+-00ADD8.svg)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/dedene/raindrop-cli)](https://goreportcard.com/report/github.com/dedene/raindrop-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A command-line interface for [Raindrop.io](https://raindrop.io) bookmark management.

## Installation

### Homebrew

```bash
brew install dedene/tap/raindrop
```

### From source

```bash
go install github.com/dedene/raindrop-cli/cmd/raindrop@latest
```

### From releases

Download the latest binary from [releases](https://github.com/dedene/raindrop-cli/releases).

## Authentication

### Test token (personal use)

Get a test token from
[raindrop.io/settings/integrations](https://raindrop.io/settings/integrations):

```bash
raindrop auth token <your-token>
raindrop auth status
```

Or use environment variable:

```bash
export RAINDROP_TOKEN=<your-token>
```

### OAuth2 (apps/shared use)

1. Create an app at [raindrop.io/app](https://raindrop.io/app).
2. Set the redirect URI to `http://localhost:8484/callback` (or update `oauth_port` in config).
3. Save credentials:

```bash
raindrop auth setup <client_id>
```

4. Authenticate:

```bash
raindrop auth login
```

## Quick Start

```bash
# Add a bookmark
raindrop add https://example.com
raindrop add https://example.com --collection Work --tags "reference,docs"

# List bookmarks
raindrop list
raindrop list --favorites
raindrop list Work --all

# Search
raindrop search "golang"
raindrop search --tag programming --type article

# Get details
raindrop get 12345

# Update
raindrop update 12345 --title "New Title" --tags "updated"

# Delete
raindrop delete 12345
```

## Commands

### Core

| Command             | Description          |
| ------------------- | -------------------- |
| `add [url]`         | Add a bookmark       |
| `list [collection]` | List bookmarks       |
| `get <id>`          | Get bookmark details |
| `update <id>`       | Update a bookmark    |
| `delete <id>`       | Delete a bookmark    |
| `search [query]`    | Search bookmarks     |

### Collections

| Command                     | Description                  |
| --------------------------- | ---------------------------- |
| `collections list`          | List collections (tree view) |
| `collections get <name>`    | Get collection details       |
| `collections create <name>` | Create a collection          |
| `collections update <name>` | Update a collection          |
| `collections delete <name>` | Delete a collection          |

### Tags

| Command                             | Description   |
| ----------------------------------- | ------------- |
| `tags list`                         | List all tags |
| `tags rename <old> <new>`           | Rename a tag  |
| `tags merge <tags> --into <target>` | Merge tags    |
| `tags delete <tags>`                | Delete tags   |

### Highlights

| Command                                 | Description        |
| --------------------------------------- | ------------------ |
| `highlights list <id>`                  | List highlights    |
| `highlights add <id> <text>`            | Add a highlight    |
| `highlights delete <id> <highlight-id>` | Delete a highlight |

### Utility

| Command                          | Description                    |
| -------------------------------- | ------------------------------ |
| `import <file>`                  | Import Netscape HTML bookmarks |
| `export --format csv\|html\|zip` | Export bookmarks               |
| `open <id>`                      | Open in browser                |
| `copy <id>`                      | Copy URL to clipboard          |

## Flags

| Flag         | Description               |
| ------------ | ------------------------- |
| `--json`     | Output JSON               |
| `--force`    | Skip confirmations        |
| `--no-input` | CI mode (fail on prompts) |
| `--verbose`  | Verbose output            |

## Shell Completions

```bash
# Bash
eval "$(raindrop completion bash)"

# Zsh
eval "$(raindrop completion zsh)"

# Fish
raindrop completion fish > ~/.config/fish/completions/raindrop.fish
```

## Configuration

Config file: `~/.config/raindrop-cli/config.yaml`

```bash
raindrop config path
raindrop config get <key>
raindrop config set <key> <value>
```

## System Collections

| Name       | ID  | Description   |
| ---------- | --- | ------------- |
| `all`      | 0   | All raindrops |
| `unsorted` | -1  | Unsorted      |
| `trash`    | -99 | Trash         |

Collection names are case-insensitive: `raindrop list Work` or `raindrop list work`.

## Bulk Operations

Add multiple URLs from stdin:

```bash
echo -e "https://a.com\nhttps://b.com" | raindrop add -
```

## License

MIT
