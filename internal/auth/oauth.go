package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// Raindrop OAuth2 endpoints.
var raindropEndpoint = oauth2.Endpoint{
	AuthURL:  "https://raindrop.io/oauth/authorize",
	TokenURL: "https://raindrop.io/oauth/access_token",
}

const defaultCallbackPort = 8484

type AuthorizeOptions struct {
	Manual  bool
	Timeout time.Duration
	Port    int
}

var (
	errAuthorization       = errors.New("authorization error")
	errMissingCode         = errors.New("missing code")
	errNoCodeInURL         = errors.New("no code found in URL")
	errNoRefreshToken      = errors.New("no refresh token received")
	errStateMismatch       = errors.New("state mismatch")
	errTokenExchangeFailed = errors.New("token exchange failed")
	errUnsupportedPlatform = errors.New("unsupported platform")
	openBrowserFn          = openBrowser
	randomStateFn          = randomState
)

func DefaultRedirectURI(port int) string {
	if port <= 0 {
		port = defaultCallbackPort
	}

	return fmt.Sprintf("http://localhost:%d/callback", port)
}

func ResolveRedirectURI(creds OAuthCredentials, port int) string {
	if strings.TrimSpace(creds.RedirectURI) != "" {
		return creds.RedirectURI
	}

	return DefaultRedirectURI(port)
}

func Authorize(ctx context.Context, creds OAuthCredentials, opts AuthorizeOptions) (string, error) {
	if opts.Timeout <= 0 {
		opts.Timeout = 2 * time.Minute
	}

	state, err := randomStateFn()
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	redirectURI := ResolveRedirectURI(creds, opts.Port)

	cfg := oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Endpoint:     raindropEndpoint,
		RedirectURL:  redirectURI,
	}

	authOpts := []oauth2.AuthCodeOption{oauth2.AccessTypeOffline}

	if opts.Manual {
		return authorizeManual(ctx, cfg, creds, state, authOpts)
	}

	return authorizeWithServer(ctx, cfg, creds, state, authOpts)
}

func authorizeManual(ctx context.Context, cfg oauth2.Config, creds OAuthCredentials, state string, authOpts []oauth2.AuthCodeOption) (string, error) {
	authURL := cfg.AuthCodeURL(state, authOpts...)

	fmt.Fprintln(os.Stderr, "Visit this URL to authorize:")
	fmt.Fprintln(os.Stderr, authURL)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "After authorizing, you'll be redirected to a URL.")
	fmt.Fprintln(os.Stderr, "Copy the URL from your browser and paste it here.")
	fmt.Fprintln(os.Stderr)
	fmt.Fprint(os.Stderr, "Paste redirect URL: ")

	var line string
	if _, err := fmt.Scanln(&line); err != nil {
		return "", fmt.Errorf("read redirect url: %w", err)
	}

	line = strings.TrimSpace(line)

	code, gotState, err := extractCodeAndState(line)
	if err != nil {
		return "", err
	}

	if gotState != "" && gotState != state {
		return "", errStateMismatch
	}

	tok, err := exchangeAuthorizationCode(ctx, creds, cfg.RedirectURL, code)
	if err != nil {
		return "", fmt.Errorf("exchange code: %w", err)
	}

	if tok.RefreshToken == "" {
		return "", errNoRefreshToken
	}

	return tok.RefreshToken, nil
}

func authorizeWithServer(ctx context.Context, cfg oauth2.Config, creds OAuthCredentials, state string, authOpts []oauth2.AuthCodeOption) (string, error) {
	parsed, err := url.Parse(cfg.RedirectURL)
	if err != nil {
		return "", fmt.Errorf("parse redirect uri: %w", err)
	}

	port := parsed.Port()
	if port == "" {
		port = fmt.Sprintf("%d", defaultCallbackPort)
	}

	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:"+port)
	if err != nil {
		return "", fmt.Errorf("listen for callback: %w", err)
	}

	defer func() { _ = ln.Close() }()

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	srv := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/callback" {
				http.NotFound(w, r)

				return
			}

			q := r.URL.Query()

			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			if q.Get("error") != "" {
				select {
				case errCh <- fmt.Errorf("%w: %s", errAuthorization, q.Get("error")):
				default:
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(cancelledHTML))

				return
			}

			if q.Get("state") != state {
				select {
				case errCh <- errStateMismatch:
				default:
				}

				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(errorHTML("State mismatch - please try again.")))

				return
			}

			code := q.Get("code")
			if code == "" {
				select {
				case errCh <- errMissingCode:
				default:
				}

				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(errorHTML("Missing authorization code.")))

				return
			}

			select {
			case codeCh <- code:
			default:
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(successHTML))
		}),
	}

	go func() {
		<-ctx.Done()

		_ = srv.Close()
	}()

	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	authURL := cfg.AuthCodeURL(state, authOpts...)

	fmt.Fprintln(os.Stderr, "Opening browser for authorization...")
	fmt.Fprintln(os.Stderr, "If the browser doesn't open, visit:")
	fmt.Fprintln(os.Stderr, authURL)
	_ = openBrowserFn(authURL)

	select {
	case code := <-codeCh:
		fmt.Fprintln(os.Stderr, "Authorization received. Finishing...")

		tok, err := exchangeAuthorizationCode(ctx, creds, cfg.RedirectURL, code)
		if err != nil {
			_ = srv.Close()

			return "", fmt.Errorf("exchange code: %w", err)
		}

		if tok.RefreshToken == "" {
			_ = srv.Close()

			return "", errNoRefreshToken
		}

		shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		_ = srv.Shutdown(shutdownCtx)

		return tok.RefreshToken, nil

	case err := <-errCh:
		_ = srv.Close()

		return "", err

	case <-ctx.Done():
		_ = srv.Close()

		return "", fmt.Errorf("authorization canceled: %w", ctx.Err())
	}
}

func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate state: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func extractCodeAndState(rawURL string) (code, state string, err error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("parse redirect url: %w", err)
	}

	code = parsed.Query().Get("code")
	if code == "" {
		return "", "", errNoCodeInURL
	}

	return code, parsed.Query().Get("state"), nil
}

func openBrowser(targetURL string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", targetURL)
	case "linux":
		cmd = exec.Command("xdg-open", targetURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", targetURL)
	default:
		return fmt.Errorf("%w: %s", errUnsupportedPlatform, runtime.GOOS)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start browser: %w", err)
	}

	return nil
}

const successHTML = `<!DOCTYPE html>
<html>
<head>
  <title>Authorization Successful</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
           display: flex; justify-content: center; align-items: center;
           min-height: 100vh; margin: 0; background: #f5f5f5; }
    .container { text-align: center; padding: 40px; background: white;
                 border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
    h1 { color: #22c55e; margin-bottom: 16px; }
    p { color: #666; }
  </style>
</head>
<body>
  <div class="container">
    <h1>&#10004; Authorization Successful</h1>
    <p>You can close this window and return to the terminal.</p>
  </div>
</body>
</html>`

const cancelledHTML = `<!DOCTYPE html>
<html>
<head>
  <title>Authorization Cancelled</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
           display: flex; justify-content: center; align-items: center;
           min-height: 100vh; margin: 0; background: #f5f5f5; }
    .container { text-align: center; padding: 40px; background: white;
                 border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
    h1 { color: #f59e0b; margin-bottom: 16px; }
    p { color: #666; }
  </style>
</head>
<body>
  <div class="container">
    <h1>Authorization Cancelled</h1>
    <p>You can close this window.</p>
  </div>
</body>
</html>`

func errorHTML(msg string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <title>Authorization Error</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
           display: flex; justify-content: center; align-items: center;
           min-height: 100vh; margin: 0; background: #f5f5f5; }
    .container { text-align: center; padding: 40px; background: white;
                 border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
    h1 { color: #ef4444; margin-bottom: 16px; }
    p { color: #666; }
  </style>
</head>
<body>
  <div class="container">
    <h1>Authorization Error</h1>
    <p>%s</p>
  </div>
</body>
</html>`, msg)
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func exchangeAuthorizationCode(ctx context.Context, creds OAuthCredentials, redirectURI, code string) (*oauth2.Token, error) {
	payload := map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"client_id":     creds.ClientID,
		"client_secret": creds.ClientSecret,
		"redirect_uri":  redirectURI,
	}

	return exchangeToken(ctx, payload)
}

func exchangeRefreshToken(ctx context.Context, creds OAuthCredentials, refreshToken string) (*oauth2.Token, error) {
	payload := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     creds.ClientID,
		"client_secret": creds.ClientSecret,
	}

	return exchangeToken(ctx, payload)
}

func exchangeToken(ctx context.Context, payload map[string]string) (*oauth2.Token, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, raindropEndpoint.TokenURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", errTokenExchangeFailed, strings.TrimSpace(string(bodyBytes)))
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	tok := &oauth2.Token{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		TokenType:    tr.TokenType,
	}

	if tr.ExpiresIn > 0 {
		tok.Expiry = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}

	return tok, nil
}
