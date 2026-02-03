package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dedene/raindrop-cli/internal/api"
)

// defaultTimeout for API calls.
const defaultTimeout = 30 * time.Second

// getClient creates an authenticated API client.
func getClient() (*api.Client, error) {
	return api.NewClientFromAuth()
}

// getClientWithContext creates client and context with timeout.
func getClientWithContext() (*api.Client, context.Context, context.CancelFunc, error) {
	client, err := getClient()
	if err != nil {
		return nil, nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)

	return client, ctx, cancel, nil
}

// confirmAction prompts for confirmation unless --force or --no-input is set.
func confirmAction(msg string, flags *RootFlags) bool {
	if flags.Force {
		return true
	}

	if flags.NoInput {
		fmt.Fprintln(os.Stderr, "confirmation required but --no-input is set")

		return false
	}

	fmt.Fprintf(os.Stderr, "%s [y/N]: ", msg)

	var response string

	_, _ = fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y" || response == "yes"
}

// readURLsFromStdin reads newline-separated URLs from stdin.
func readURLsFromStdin() ([]string, error) {
	var urls []string

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			urls = append(urls, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}

	return urls, nil
}

// truncate shortens a string to maxLen with ellipsis.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	if maxLen <= 3 {
		return s[:maxLen]
	}

	return s[:maxLen-3] + "..."
}
