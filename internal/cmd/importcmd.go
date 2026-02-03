package cmd

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dedene/raindrop-cli/internal/errfmt"
)

type ImportCmd struct {
	File string `arg:"" help:"Netscape bookmark HTML file" type:"existingfile"`
}

func (c *ImportCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	// Read the file
	data, err := os.ReadFile(c.File)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("import", filepath.Base(c.File))
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}

	if _, copyErr := io.Copy(part, bytes.NewReader(data)); copyErr != nil {
		return fmt.Errorf("write file: %w", copyErr)
	}

	if closeErr := writer.Close(); closeErr != nil {
		return fmt.Errorf("close writer: %w", closeErr)
	}

	// Make request using raw HTTP since we need multipart
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.raindrop.io/rest/v1/import/file", body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	tok, err := client.Token(ctx)
	if err != nil {
		return errfmt.Format(err)
	}

	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("import failed: %s - %s", resp.Status, string(respBody))
	}

	fmt.Fprintln(os.Stdout, "Import started successfully.")
	fmt.Fprintln(os.Stdout, "Check raindrop.io for import progress.")

	return nil
}
