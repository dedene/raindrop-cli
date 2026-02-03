package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/oauth2"

	"github.com/dedene/raindrop-cli/internal/auth"
)

const (
	BaseURL     = "https://api.raindrop.io/rest/v1"
	UserAgent   = "raindrop-cli/0.1.0"
	ContentType = "application/json"
)

// Client is the Raindrop.io API client.
type Client struct {
	baseURL     string
	httpClient  *http.Client
	tokenSource oauth2.TokenSource
}

// NewClient creates a new API client with the given token source.
func NewClient(ts oauth2.TokenSource) *Client {
	return &Client{
		baseURL:     BaseURL,
		tokenSource: ts,
		httpClient: &http.Client{
			Transport: NewRetryTransport(http.DefaultTransport),
		},
	}
}

// NewClientWithBaseURL creates a new API client with a custom base URL.
func NewClientWithBaseURL(ts oauth2.TokenSource, baseURL string) *Client {
	client := NewClient(ts)
	if baseURL != "" {
		client.baseURL = baseURL
	}

	return client
}

// Token returns the current OAuth token.
func (c *Client) Token(ctx context.Context) (*oauth2.Token, error) {
	_ = ctx // context available for future use

	tok, err := c.tokenSource.Token()
	if err != nil {
		return nil, &AuthError{Err: err}
	}

	return tok, nil
}

// NewClientFromAuth creates a client using stored auth credentials.
func NewClientFromAuth() (*Client, error) {
	// Check environment variable first
	if envToken := os.Getenv("RAINDROP_TOKEN"); envToken != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: envToken,
		})

		return NewClient(ts), nil
	}

	store, err := auth.OpenDefault()
	if err != nil {
		return nil, fmt.Errorf("open keyring: %w", err)
	}

	tok, err := store.GetToken()
	if err != nil {
		if errors.Is(err, auth.ErrNoToken) {
			return nil, auth.ErrNotAuthenticated
		}

		return nil, fmt.Errorf("get token: %w", err)
	}

	if tok.TestToken != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: tok.TestToken,
		})

		return NewClient(ts), nil
	}

	if tok.RefreshToken == "" {
		return nil, auth.ErrNotAuthenticated
	}

	creds, err := store.GetCredentials()
	if err != nil {
		return nil, fmt.Errorf("get credentials: %w", err)
	}

	ts := auth.NewOAuthTokenSource(store, creds)

	return NewClient(ts), nil
}

func (c *Client) do(ctx context.Context, method, path string, body []byte, out interface{}) error {
	reqURL := c.baseURL + path

	for attempt := 0; attempt < 2; attempt++ {
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		tok, err := c.tokenSource.Token()
		if err != nil {
			return &AuthError{Err: err}
		}

		req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set("Accept", ContentType)

		if body != nil {
			req.Header.Set("Content-Type", ContentType)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("do request: %w", err)
		}

		if resp.StatusCode == http.StatusUnauthorized {
			if ts, ok := c.tokenSource.(interface{ Invalidate() }); ok {
				ts.Invalidate()
			}

			drainAndClose(resp.Body)

			if attempt == 0 {
				continue
			}

			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    "unauthorized",
				Details:    "token may be invalid or expired; check your token at raindrop.io/settings/integrations",
			}
		}

		if resp.StatusCode == http.StatusNotFound {
			defer resp.Body.Close()

			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    "not found",
			}
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			defer resp.Body.Close()

			return &RateLimitError{}
		}

		if resp.StatusCode >= 400 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()

			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    http.StatusText(resp.StatusCode),
				Details:    string(bodyBytes),
			}
		}

		if out != nil && resp.StatusCode != http.StatusNoContent {
			defer resp.Body.Close()

			if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}

			return nil
		}

		_ = resp.Body.Close()

		return nil
	}

	return &APIError{
		StatusCode: http.StatusUnauthorized,
		Message:    "unauthorized",
		Details:    "token may be invalid or expired; check your token at raindrop.io/settings/integrations",
	}
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, out interface{}) error {
	return c.do(ctx, http.MethodGet, path, nil, out)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body interface{}, out interface{}) error {
	var bodyBytes []byte

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}

		bodyBytes = data
	}

	return c.do(ctx, http.MethodPost, path, bodyBytes, out)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, body interface{}, out interface{}) error {
	var bodyBytes []byte

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}

		bodyBytes = data
	}

	return c.do(ctx, http.MethodPut, path, bodyBytes, out)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

// GetUser returns the authenticated user's info.
func (c *Client) GetUser(ctx context.Context) (*User, error) {
	var resp struct {
		User User `json:"user"`
	}

	if err := c.Get(ctx, "/user", &resp); err != nil {
		return nil, err
	}

	return &resp.User, nil
}
