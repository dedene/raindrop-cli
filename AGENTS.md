# Repository Guidelines

## Project Structure

- `cmd/raindrop/`: CLI entrypoint
- `internal/`: implementation
  - `cmd/`: command routing (kong CLI framework)
  - `api/`: Raindrop.io HTTP client
  - `auth/`: OAuth2 token management + keyring storage
  - `config/`: YAML config (timezone, output format, OAuth port, hyperlinks)
  - `output/`: formatters (table, JSON, hyperlinks)
  - `errfmt/`: error formatting
- `bin/`: build outputs

## Build, Test, and Development Commands

- `make build`: compile to `bin/raindrop`
- `make fmt` / `make lint` / `make test` / `make ci`: format, lint, test, full local gate
- `make tools`: install pinned dev tools into `.tools/`
- `make clean`: remove bin/ and .tools/

## Coding Style & Naming Conventions

- Formatting: `make fmt` (goimports local prefix `github.com/dedene/raindrop-cli` + gofumpt)
- Output: keep stdout parseable (`--json`); send human hints/progress to stderr
- Linting: golangci-lint v2.1.6 with project config
- Shell completion supported

## Testing Guidelines

- Unit tests: stdlib `testing`
- **Note**: No test files currently present (area for improvement)
- CI gate: fmt-check, lint, test

## Config & Secrets

- **OAuth2**: full OAuth flow for Raindrop.io authentication
- **Keyring**: 99designs/keyring for token storage + caching
- **Config file**: `~/.config/raindrop/config.yaml`
  - `timezone`: display timezone
  - `default_output`: output format preference
  - `oauth_port`: local OAuth callback port
  - `hyperlinks`: terminal hyperlink mode

## Key Commands

- `add/list/get/update/delete`: bookmark CRUD operations
- `search`: full-text search across bookmarks
- `collections`: manage collections
- `tags`: manage tags
- `highlights`: manage highlights
- `import/export`: bulk operations
- `auth`: authentication flow
- Global flags: `--json`, `--verbose`, `--force`, `--no-input`, `--hyperlinks`

## Commit & Pull Request Guidelines

- Conventional Commits: `feat|fix|refactor|build|ci|chore|docs|style|perf|test`
- Group related changes; avoid bundling unrelated refactors
- PR review: use `gh pr view` / `gh pr diff`; don't switch branches

## Security Tips

- Never commit OAuth tokens or client secrets
- Prefer OS keychain for token storage
- OAuth callback runs on localhost only
