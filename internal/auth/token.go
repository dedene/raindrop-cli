package auth

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// TestTokenSource provides a static token from the test token or environment.
type TestTokenSource struct {
	mu    sync.Mutex
	store Store
}

func NewTestTokenSource(store Store) *TestTokenSource {
	return &TestTokenSource{
		store: store,
	}
}

// Token returns the access token from env or keyring.
func (ts *TestTokenSource) Token() (*oauth2.Token, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Check environment variable first
	if envToken := os.Getenv("RAINDROP_TOKEN"); envToken != "" {
		return &oauth2.Token{
			AccessToken: envToken,
		}, nil
	}

	// Get from keyring
	tok, err := ts.store.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	if tok.TestToken != "" {
		return &oauth2.Token{
			AccessToken: tok.TestToken,
		}, nil
	}

	return nil, ErrNotAuthenticated
}

// RefreshTokenSource uses a refresh token directly (used during login).
type RefreshTokenSource struct {
	creds        OAuthCredentials
	refreshToken string
}

// NewRefreshTokenSource creates a token source from a refresh token directly.
func NewRefreshTokenSource(creds OAuthCredentials, refreshToken string) *RefreshTokenSource {
	return &RefreshTokenSource{
		creds:        creds,
		refreshToken: refreshToken,
	}
}

// Token returns an access token by exchanging the refresh token.
func (ts *RefreshTokenSource) Token() (*oauth2.Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tok, err := exchangeRefreshToken(ctx, ts.creds, ts.refreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}

	return tok, nil
}

// OAuthTokenSource provides OAuth2 tokens with lazy refresh.
// Access tokens are kept in memory only; refresh tokens are stored in keyring.
type OAuthTokenSource struct {
	mu           sync.Mutex
	store        Store
	creds        OAuthCredentials
	accessToken  string
	accessExpiry time.Time
}

// NewOAuthTokenSource creates an OAuth token source using stored credentials.
func NewOAuthTokenSource(store Store, creds OAuthCredentials) *OAuthTokenSource {
	return &OAuthTokenSource{
		store: store,
		creds: creds,
	}
}

// Token returns a valid access token, refreshing if necessary.
func (ts *OAuthTokenSource) Token() (*oauth2.Token, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.accessToken != "" && !ts.accessExpiry.IsZero() && time.Now().Before(ts.accessExpiry) {
		return &oauth2.Token{
			AccessToken: ts.accessToken,
			Expiry:      ts.accessExpiry,
		}, nil
	}

	if err := ts.refresh(); err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: ts.accessToken,
		Expiry:      ts.accessExpiry,
	}, nil
}

// Invalidate marks the current access token as invalid, forcing a refresh on next Token() call.
func (ts *OAuthTokenSource) Invalidate() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.accessToken = ""
	ts.accessExpiry = time.Time{}
}

func (ts *OAuthTokenSource) refresh() error {
	tok, err := ts.store.GetToken()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNotAuthenticated, err)
	}

	if tok.RefreshToken == "" {
		return ErrNotAuthenticated
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	newTok, err := exchangeRefreshToken(ctx, ts.creds, tok.RefreshToken)
	if err != nil {
		return fmt.Errorf("refresh token: %w", err)
	}

	ts.accessToken = newTok.AccessToken
	ts.accessExpiry = newTok.Expiry

	if newTok.RefreshToken != "" && newTok.RefreshToken != tok.RefreshToken {
		tok.RefreshToken = newTok.RefreshToken
		if err := ts.store.SetToken(tok); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to store new refresh token: %v\n", err)
		}
	}

	return nil
}
