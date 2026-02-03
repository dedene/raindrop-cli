package cmd

import (
	"fmt"
	"os"
)

type CompletionCmd struct {
	Bash BashCompletionCmd `cmd:"" help:"Generate bash completion script"`
	Zsh  ZshCompletionCmd  `cmd:"" help:"Generate zsh completion script"`
	Fish FishCompletionCmd `cmd:"" help:"Generate fish completion script"`
}

type BashCompletionCmd struct{}

func (c *BashCompletionCmd) Run() error {
	script := `_raindrop_completions() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Main commands
    local commands="add list get update delete search collections tags highlights import export open copy auth config version completion"

    # Subcommands
    local auth_cmds="setup login token status logout"
    local config_cmds="path get set"
    local collections_cmds="list get create update delete"
    local tags_cmds="list rename merge delete"
    local highlights_cmds="list add delete"
    local completion_cmds="bash zsh fish"

    case "$prev" in
        raindrop)
            COMPREPLY=($(compgen -W "$commands" -- "$cur"))
            return
            ;;
        auth)
            COMPREPLY=($(compgen -W "$auth_cmds" -- "$cur"))
            return
            ;;
        config)
            COMPREPLY=($(compgen -W "$config_cmds" -- "$cur"))
            return
            ;;
        collections)
            COMPREPLY=($(compgen -W "$collections_cmds" -- "$cur"))
            return
            ;;
        tags)
            COMPREPLY=($(compgen -W "$tags_cmds" -- "$cur"))
            return
            ;;
        highlights)
            COMPREPLY=($(compgen -W "$highlights_cmds" -- "$cur"))
            return
            ;;
        completion)
            COMPREPLY=($(compgen -W "$completion_cmds" -- "$cur"))
            return
            ;;
    esac

    # Handle flags
    if [[ "$cur" == -* ]]; then
        local flags="--help --json --verbose --force --no-input --version"
        COMPREPLY=($(compgen -W "$flags" -- "$cur"))
        return
    fi
}

complete -F _raindrop_completions raindrop
`
	fmt.Fprintln(os.Stdout, script)
	fmt.Fprintln(os.Stderr, "# Add this to ~/.bashrc:")
	fmt.Fprintln(os.Stderr, "# eval \"$(raindrop completion bash)\"")

	return nil
}

type ZshCompletionCmd struct{}

func (c *ZshCompletionCmd) Run() error {
	script := `#compdef raindrop

_raindrop() {
    local -a commands
    commands=(
        'add:Add a bookmark'
        'list:List bookmarks'
        'get:Get bookmark details'
        'update:Update a bookmark'
        'delete:Delete a bookmark'
        'search:Search bookmarks'
        'collections:Manage collections'
        'tags:Manage tags'
        'highlights:Manage highlights'
        'import:Import bookmarks from HTML file'
        'export:Export bookmarks'
        'open:Open bookmark in browser'
        'copy:Copy bookmark URL to clipboard'
        'auth:Authentication and credentials'
        'config:Manage configuration'
        'version:Print version'
        'completion:Generate shell completions'
    )

    _arguments -C \
        '--help[Show help]' \
        '--json[Output JSON]' \
        '--verbose[Enable verbose logging]' \
        '--force[Skip confirmations]' \
        '--no-input[Fail instead of prompting]' \
        '--version[Print version]' \
        '1: :->cmd' \
        '*::arg:->args'

    case "$state" in
        cmd)
            _describe -t commands 'raindrop commands' commands
            ;;
        args)
            case $words[1] in
                auth)
                    local -a auth_cmds
                    auth_cmds=('setup:Configure OAuth credentials' 'login:OAuth login' 'token:Set test token' 'status:Show auth status' 'logout:Remove credentials')
                    _describe -t commands 'auth commands' auth_cmds
                    ;;
                collections)
                    local -a col_cmds
                    col_cmds=('list:List collections' 'get:Get collection' 'create:Create collection' 'update:Update collection' 'delete:Delete collection')
                    _describe -t commands 'collection commands' col_cmds
                    ;;
                tags)
                    local -a tag_cmds
                    tag_cmds=('list:List tags' 'rename:Rename tag' 'merge:Merge tags' 'delete:Delete tags')
                    _describe -t commands 'tag commands' tag_cmds
                    ;;
                highlights)
                    local -a hl_cmds
                    hl_cmds=('list:List highlights' 'add:Add highlight' 'delete:Delete highlight')
                    _describe -t commands 'highlight commands' hl_cmds
                    ;;
            esac
            ;;
    esac
}

_raindrop "$@"
`
	fmt.Fprintln(os.Stdout, script)
	fmt.Fprintln(os.Stderr, "# Add this to ~/.zshrc:")
	fmt.Fprintln(os.Stderr, "# eval \"$(raindrop completion zsh)\"")

	return nil
}

type FishCompletionCmd struct{}

func (c *FishCompletionCmd) Run() error {
	script := `# Fish completion for raindrop

# Disable file completion by default
complete -c raindrop -f

# Main commands
complete -c raindrop -n "__fish_use_subcommand" -a "add" -d "Add a bookmark"
complete -c raindrop -n "__fish_use_subcommand" -a "list" -d "List bookmarks"
complete -c raindrop -n "__fish_use_subcommand" -a "get" -d "Get bookmark details"
complete -c raindrop -n "__fish_use_subcommand" -a "update" -d "Update a bookmark"
complete -c raindrop -n "__fish_use_subcommand" -a "delete" -d "Delete a bookmark"
complete -c raindrop -n "__fish_use_subcommand" -a "search" -d "Search bookmarks"
complete -c raindrop -n "__fish_use_subcommand" -a "collections" -d "Manage collections"
complete -c raindrop -n "__fish_use_subcommand" -a "tags" -d "Manage tags"
complete -c raindrop -n "__fish_use_subcommand" -a "highlights" -d "Manage highlights"
complete -c raindrop -n "__fish_use_subcommand" -a "import" -d "Import bookmarks"
complete -c raindrop -n "__fish_use_subcommand" -a "export" -d "Export bookmarks"
complete -c raindrop -n "__fish_use_subcommand" -a "open" -d "Open in browser"
complete -c raindrop -n "__fish_use_subcommand" -a "copy" -d "Copy URL"
complete -c raindrop -n "__fish_use_subcommand" -a "auth" -d "Authentication"
complete -c raindrop -n "__fish_use_subcommand" -a "config" -d "Configuration"
complete -c raindrop -n "__fish_use_subcommand" -a "version" -d "Print version"
complete -c raindrop -n "__fish_use_subcommand" -a "completion" -d "Shell completions"

# Auth subcommands
complete -c raindrop -n "__fish_seen_subcommand_from auth" -a "token" -d "Set test token"
complete -c raindrop -n "__fish_seen_subcommand_from auth" -a "setup" -d "Configure OAuth credentials"
complete -c raindrop -n "__fish_seen_subcommand_from auth" -a "login" -d "OAuth login"
complete -c raindrop -n "__fish_seen_subcommand_from auth" -a "status" -d "Show status"
complete -c raindrop -n "__fish_seen_subcommand_from auth" -a "logout" -d "Logout"

# Collections subcommands
complete -c raindrop -n "__fish_seen_subcommand_from collections" -a "list" -d "List"
complete -c raindrop -n "__fish_seen_subcommand_from collections" -a "get" -d "Get"
complete -c raindrop -n "__fish_seen_subcommand_from collections" -a "create" -d "Create"
complete -c raindrop -n "__fish_seen_subcommand_from collections" -a "update" -d "Update"
complete -c raindrop -n "__fish_seen_subcommand_from collections" -a "delete" -d "Delete"

# Tags subcommands
complete -c raindrop -n "__fish_seen_subcommand_from tags" -a "list" -d "List"
complete -c raindrop -n "__fish_seen_subcommand_from tags" -a "rename" -d "Rename"
complete -c raindrop -n "__fish_seen_subcommand_from tags" -a "merge" -d "Merge"
complete -c raindrop -n "__fish_seen_subcommand_from tags" -a "delete" -d "Delete"

# Global flags
complete -c raindrop -l help -d "Show help"
complete -c raindrop -l json -d "Output JSON"
complete -c raindrop -l verbose -d "Verbose logging"
complete -c raindrop -l force -d "Skip confirmations"
complete -c raindrop -l no-input -d "Non-interactive mode"
complete -c raindrop -l version -d "Print version"
`
	fmt.Fprintln(os.Stdout, script)
	fmt.Fprintln(os.Stderr, "# Save to ~/.config/fish/completions/raindrop.fish:")
	fmt.Fprintln(os.Stderr, "# raindrop completion fish > ~/.config/fish/completions/raindrop.fish")

	return nil
}
