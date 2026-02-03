package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dedene/raindrop-cli/internal/errfmt"
)

type ExportCmd struct {
	Collection string `arg:"" optional:"" help:"Collection name/ID (default: all)" default:"0"`
	Format     string `required:"" help:"Export format (csv, html, zip)" enum:"csv,html,zip" short:"f"`
	Output     string `help:"Output file (default: stdout)" short:"o"`
}

func (c *ExportCmd) Run(flags *RootFlags) error {
	// Zip format requires output file (binary data)
	if c.Format == "zip" && c.Output == "" {
		return fmt.Errorf("zip format requires --output file (binary data cannot be written to stdout)")
	}

	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	collectionID, err := client.ResolveCollection(ctx, c.Collection)
	if err != nil {
		return errfmt.Format(err)
	}

	// Build export URL
	exportURL := fmt.Sprintf("https://api.raindrop.io/rest/v1/raindrops/%d/export.%s", collectionID, c.Format)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, exportURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	tok, err := client.Token(ctx)
	if err != nil {
		return errfmt.Format(err)
	}

	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("export failed: %s - %s", resp.Status, string(respBody))
	}

	// Determine output destination
	var out io.Writer = os.Stdout

	if c.Output != "" {
		f, createErr := os.Create(c.Output)
		if createErr != nil {
			return fmt.Errorf("create output file: %w", createErr)
		}
		defer f.Close()

		out = f
	}

	// Copy response to output
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	if c.Output != "" {
		fmt.Fprintf(os.Stderr, "Exported to %s (%d bytes)\n", c.Output, written)
	}

	return nil
}
