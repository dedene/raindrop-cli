package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/dedene/raindrop-cli/internal/api"
	"github.com/dedene/raindrop-cli/internal/auth"
	"github.com/dedene/raindrop-cli/internal/config"
	"github.com/dedene/raindrop-cli/internal/errfmt"
)

type AuthCmd struct {
	Setup  AuthSetupCmd  `cmd:"" help:"Configure OAuth client credentials"`
	Token  AuthTokenCmd  `cmd:"" help:"Set test token for authentication"`
	Login  AuthLoginCmd  `cmd:"" help:"Authenticate with OAuth"`
	Status AuthStatusCmd `cmd:"" help:"Show authentication status"`
	Logout AuthLogoutCmd `cmd:"" help:"Remove stored tokens"`
}

type AuthSetupCmd struct {
	ClientID     string `arg:"" help:"OAuth client ID"`
	ClientSecret string `help:"OAuth client secret (omit to prompt securely)"`
	RedirectURI  string `help:"OAuth redirect URI (default: http://localhost:<oauth_port>/callback)"`
}

func (c *AuthSetupCmd) Run() error {
	if c.ClientID == "" {
		msg := "expected \"<client-id>\"\n\n" +
			"Create a developer app at https://app.raindrop.io/settings/integrations\n" +
			"Use http://localhost:8484/callback as Redirect URI\n" +
			"Change port via: raindrop config set oauth_port <port>\n\n" +
			"Example:\n  raindrop auth setup <client_id>"

		return &ExitError{Code: ExitUsage, Err: errors.New(msg)}
	}

	clientSecret := c.ClientSecret
	if clientSecret == "" {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return errors.New("no TTY available; use --client-secret flag")
		}

		fmt.Fprint(os.Stderr, "Enter client secret: ")

		secretBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)

		if err != nil {
			return fmt.Errorf("read secret: %w", err)
		}

		clientSecret = strings.TrimSpace(string(secretBytes))
	}

	if clientSecret == "" {
		return errors.New("client secret cannot be empty")
	}

	store, err := auth.OpenDefault()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	cfg, err := config.ReadConfig()
	if err != nil {
		return err
	}

	redirectURI := c.RedirectURI
	if redirectURI == "" {
		redirectURI = auth.DefaultRedirectURI(cfg.OAuthPort)
	}

	creds := auth.OAuthCredentials{
		ClientID:     c.ClientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
	}

	if err := store.SetCredentials(creds); err != nil {
		return fmt.Errorf("store credentials: %w", err)
	}

	fmt.Fprintln(os.Stdout, "OAuth credentials saved.")
	fmt.Fprintln(os.Stdout, "Run 'raindrop auth login' to authenticate.")

	return nil
}

type AuthTokenCmd struct {
	Token string `arg:"" help:"Test token from raindrop.io/settings/integrations"`
}

func (c *AuthTokenCmd) Run() error {
	if c.Token == "" {
		return fmt.Errorf("token is required")
	}

	store, err := auth.OpenDefault()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	tok := auth.Token{
		TestToken:    c.Token,
		RefreshToken: "",
		CreatedAt:    time.Now().UTC(),
	}

	if err := store.SetToken(tok); err != nil {
		return fmt.Errorf("store token: %w", err)
	}

	fmt.Fprintln(os.Stdout, "Token saved successfully.")
	fmt.Fprintln(os.Stdout, "Run 'raindrop auth status' to verify.")

	return nil
}

type AuthLoginCmd struct {
	Manual bool `help:"Manual authorization (paste URL instead of callback server)"`
}

func (c *AuthLoginCmd) Run() error {
	store, err := auth.OpenDefault()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	creds, err := store.GetCredentials()
	if err != nil {
		if errors.Is(err, auth.ErrNoCredentials) {
			return fmt.Errorf("OAuth credentials not configured\n  Run: raindrop auth setup <client_id>")
		}

		return fmt.Errorf("read credentials: %w", err)
	}

	cfg, err := config.ReadConfig()
	if err != nil {
		return err
	}

	ctx := context.Background()
	refreshToken, err := auth.Authorize(ctx, creds, auth.AuthorizeOptions{
		Manual:  c.Manual,
		Timeout: 3 * time.Minute,
		Port:    cfg.OAuthPort,
	})
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}

	tok := auth.Token{
		TestToken:    "",
		RefreshToken: refreshToken,
		CreatedAt:    time.Now().UTC(),
	}

	if storeErr := store.SetToken(tok); storeErr != nil {
		return fmt.Errorf("store token: %w", storeErr)
	}

	ts := auth.NewRefreshTokenSource(creds, refreshToken)
	client := api.NewClient(ts)
	user, fetchErr := c.fetchUser(ctx, client)
	if fetchErr != nil {
		fmt.Fprintln(os.Stdout, "Authenticated successfully.")

		return nil //nolint:nilerr // user fetch is optional; auth succeeded
	}

	fmt.Fprintf(os.Stdout, "Successfully authenticated as %s\n", user.FullName)

	return nil
}

func (c *AuthLoginCmd) fetchUser(ctx context.Context, client *api.Client) (*api.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return client.GetUser(ctx)
}

type AuthStatusCmd struct{}

func (c *AuthStatusCmd) Run() error {
	// Check environment variable first
	if os.Getenv("RAINDROP_TOKEN") != "" {
		fmt.Fprintln(os.Stdout, "Using token from RAINDROP_TOKEN environment variable")

		return c.verifyToken()
	}

	store, err := auth.OpenDefault()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	credsConfigured, err := store.CredentialsExists()
	if err != nil {
		return fmt.Errorf("check credentials: %w", err)
	}

	tok, err := store.GetToken()
	if err != nil {
		if credsConfigured {
			fmt.Fprintln(os.Stdout, "OAuth credentials configured but not authenticated.")
			fmt.Fprintln(os.Stdout, "Run 'raindrop auth login' to authenticate.")
		} else {
			fmt.Fprintln(os.Stdout, "Not authenticated")
			fmt.Fprintln(os.Stdout, "Run 'raindrop auth token <token>' or 'raindrop auth login' to authenticate.")
		}

		return nil
	}

	if tok.TestToken != "" {
		fmt.Fprintf(os.Stdout, "Authenticated with test token (since %s)\n", tok.CreatedAt.Format("2006-01-02"))

		return c.verifyToken()
	}

	if tok.RefreshToken != "" {
		if !credsConfigured {
			fmt.Fprintln(os.Stdout, "OAuth token stored but client not configured.")
			fmt.Fprintln(os.Stdout, "Run 'raindrop auth setup <client_id>'.")

			return nil
		}

		fmt.Fprintf(os.Stdout, "Authenticated with OAuth (since %s)\n", tok.CreatedAt.Format("2006-01-02"))

		return c.verifyToken()
	}

	if credsConfigured {
		fmt.Fprintln(os.Stdout, "OAuth credentials configured but not authenticated.")
		fmt.Fprintln(os.Stdout, "Run 'raindrop auth login' to authenticate.")

		return nil
	}

	fmt.Fprintln(os.Stdout, "Not authenticated")
	fmt.Fprintln(os.Stdout, "Run 'raindrop auth token <token>' or 'raindrop auth login' to authenticate.")

	return nil
}

func (c *AuthStatusCmd) verifyToken() error {
	client, err := api.NewClientFromAuth()
	if err != nil {
		return errfmt.Format(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := client.GetUser(ctx)
	if err != nil {
		return errfmt.Format(err)
	}

	fmt.Fprintf(os.Stdout, "User: %s\n", user.FullName)

	if user.Pro {
		fmt.Fprintln(os.Stdout, "Plan: PRO")
	} else {
		fmt.Fprintln(os.Stdout, "Plan: Free")
	}

	return nil
}

type AuthLogoutCmd struct{}

func (c *AuthLogoutCmd) Run() error {
	store, err := auth.OpenDefault()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	if err := store.DeleteToken(); err != nil {
		return fmt.Errorf("remove token: %w", err)
	}

	fmt.Fprintln(os.Stdout, "Logged out successfully.")

	return nil
}
