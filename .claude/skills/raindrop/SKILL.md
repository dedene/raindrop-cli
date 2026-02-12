---
name: raindrop-cli
description: >
  Manage Raindrop.io bookmarks via the raindrop CLI. Use when the user wants to add, list,
  search, update, or delete bookmarks, organize collections, manage tags, work with highlights,
  import/export bookmarks, or automate bookmark workflows. Triggered by mentions of Raindrop,
  bookmarks, reading lists, collections, or link management.
---

# raindrop-cli

Command-line interface for [Raindrop.io](https://raindrop.io) bookmark management. Add, search, organize, and export bookmarks.

## Quick Start

```bash
# Verify auth
raindrop auth status

# Add a bookmark
raindrop add https://example.com --tags "reference"

# List bookmarks
raindrop list --json

# Search
raindrop search "golang" --json
```

## Authentication

### Test Token (personal use)

Get a test token from [raindrop.io/settings/integrations](https://raindrop.io/settings/integrations):

```bash
raindrop auth token <your-token>
raindrop auth status
```

Or use environment variable:

```bash
export RAINDROP_TOKEN=<your-token>
```

### OAuth2 (apps/shared use)

Requires browser-based OAuth flow. User must run `raindrop auth login` interactively. Do not attempt OAuth setup on behalf of the user.

## Core Rules

1. **Always use `--json`** when parsing output. Table format is for display only.
2. **Read before write** -- fetch bookmark/collection before modifying.
3. **Collection names are case-insensitive** -- `Work` and `work` are equivalent.
4. **Use `--force`** to skip confirmations in scripts.
5. **Pipe with jq** -- extract IDs: `raindrop list --json | jq -r '.[].id'`

## Output Formats

| Flag | Format | Use case |
|------|--------|----------|
| (default) | Table | User-facing display |
| `--json` | JSON | Agent parsing, scripting |

## System Collections

| Name | ID | Description |
|------|-----|-------------|
| `all` | 0 | All bookmarks |
| `unsorted` | -1 | Unsorted bookmarks |
| `trash` | -99 | Trash |

## Workflows

### Add Bookmarks

```bash
# Basic add
raindrop add https://example.com

# With metadata
raindrop add https://example.com --collection Work --tags "docs,reference"

# Bulk add from stdin
echo -e "https://a.com\nhttps://b.com" | raindrop add -
```

### List and Search

```bash
# List all bookmarks
raindrop list --json

# List from collection
raindrop list Work --json
raindrop list Work --all --json  # include nested

# Favorites only
raindrop list --favorites --json

# Search by text
raindrop search "golang tutorial" --json

# Search with filters
raindrop search --tag programming --type article --json
```

### Get and Update

```bash
# Get bookmark details
raindrop get 12345 --json

# Update bookmark
raindrop update 12345 --title "New Title" --tags "updated,important"

# Open in browser
raindrop open 12345

# Copy URL to clipboard
raindrop copy 12345
```

### Delete

```bash
# Delete with confirmation
raindrop delete 12345

# Delete without confirmation
raindrop delete 12345 --force
```

### Collections

```bash
# List collections (tree view)
raindrop collections list --json

# Get collection details
raindrop collections get Work --json

# Create collection
raindrop collections create "New Project"

# Update collection
raindrop collections update Work --title "Work Projects"

# Delete collection
raindrop collections delete "Old Project" --force
```

### Tags

```bash
# List all tags
raindrop tags list --json

# Rename a tag
raindrop tags rename old-name new-name

# Merge tags
raindrop tags merge tag1 tag2 --into combined

# Delete tags
raindrop tags delete unused-tag
```

### Highlights

```bash
# List highlights for a bookmark
raindrop highlights list 12345 --json

# Add a highlight
raindrop highlights add 12345 "Important quote from the page"

# Delete a highlight
raindrop highlights delete 12345 <highlight-id>
```

### Import/Export

```bash
# Import Netscape HTML bookmarks
raindrop import bookmarks.html

# Export all bookmarks
raindrop export --format csv > bookmarks.csv
raindrop export --format html > bookmarks.html
raindrop export --format zip > backup.zip
```

## Scripting Examples

```bash
# Get bookmark ID after adding
ID=$(raindrop add https://example.com --json | jq -r '.id')

# List all bookmark URLs
raindrop list --json | jq -r '.[].link'

# Find bookmarks without tags
raindrop list --json | jq -r '.[] | select(.tags | length == 0) | .link'

# Count bookmarks per collection
raindrop collections list --json | jq -r '.[] | "\(.title): \(.count)"'

# Delete all trash
raindrop list trash --json | jq -r '.[].id' | xargs -I{} raindrop delete {} --force
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `RAINDROP_TOKEN` | API token (skips keyring) |

## Command Reference

| Command | Description |
|---------|-------------|
| `add [url]` | Add a bookmark |
| `list [collection]` | List bookmarks |
| `get <id>` | Get bookmark details |
| `update <id>` | Update a bookmark |
| `delete <id>` | Delete a bookmark |
| `search [query]` | Search bookmarks |
| `open <id>` | Open in browser |
| `copy <id>` | Copy URL to clipboard |
| `collections list` | List collections |
| `collections get <name>` | Get collection |
| `collections create <name>` | Create collection |
| `collections update <name>` | Update collection |
| `collections delete <name>` | Delete collection |
| `tags list` | List all tags |
| `tags rename <old> <new>` | Rename tag |
| `tags merge <tags> --into <target>` | Merge tags |
| `tags delete <tags>` | Delete tags |
| `highlights list <id>` | List highlights |
| `highlights add <id> <text>` | Add highlight |
| `highlights delete <id> <hid>` | Delete highlight |
| `import <file>` | Import bookmarks |
| `export` | Export bookmarks |
| `auth token <token>` | Set API token |
| `auth status` | Check auth status |
| `config get <key>` | Get config value |
| `config set <key> <value>` | Set config value |

## Guidelines

- Never expose or log API tokens.
- Confirm destructive operations (`delete`) with the user first.
- Collections and tags are case-insensitive for matching.
- Use `--no-input` in CI/scripts to fail on prompts instead of hanging.


## Installation

```bash
brew install dedene/tap/raindrop
```
