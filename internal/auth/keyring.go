package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"golang.org/x/term"

	"github.com/dedene/raindrop-cli/internal/config"
)

type Store interface {
	SetToken(tok Token) error
	GetToken() (Token, error)
	DeleteToken() error
	SetCredentials(creds OAuthCredentials) error
	GetCredentials() (OAuthCredentials, error)
	DeleteCredentials() error
	CredentialsExists() (bool, error)
}

type KeyringStore struct {
	ring        keyring.Keyring
	tokenCache  *Token
	credsCache  *OAuthCredentials
	tokenLoaded bool
	credsLoaded bool
}

type Token struct {
	TestToken    string    `json:"test_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
}

type OAuthCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
}

const (
	keyringPasswordEnv = "RAINDROP_KEYRING_PASSWORD" //nolint:gosec // env var name
	keyringBackendEnv  = "RAINDROP_KEYRING_BACKEND"  //nolint:gosec // env var name
	tokenKey           = "token"
	credentialsKey     = "oauth_credentials" //nolint:gosec // keyring key name, not a credential
)

var (
	// ErrNoToken is returned when no token is stored in the keyring.
	ErrNoToken             = errors.New("no token found")
	ErrNoCredentials       = errors.New("no oauth credentials configured")
	errNoTTY               = errors.New("no TTY available for keyring password prompt")
	errInvalidBackend      = errors.New("invalid keyring backend")
	errKeyringTimeout      = errors.New("keyring connection timed out")
	openKeyringFunc        = openKeyring
	keyringOpenFunc        = keyring.Open
	errMissingClientID     = errors.New("missing client id")
	errMissingClientSecret = errors.New("missing client secret")
)

const keyringOpenTimeout = 5 * time.Second

func openKeyring() (keyring.Keyring, error) {
	keyringDir, err := config.EnsureKeyringDir()
	if err != nil {
		return nil, fmt.Errorf("ensure keyring dir: %w", err)
	}

	backend := normalizeBackend(os.Getenv(keyringBackendEnv))

	backends, err := allowedBackends(backend)
	if err != nil {
		return nil, err
	}

	dbusAddr := os.Getenv("DBUS_SESSION_BUS_ADDRESS")
	if shouldForceFileBackend(runtime.GOOS, backend, dbusAddr) {
		backends = []keyring.BackendType{keyring.FileBackend}
	}

	cfg := keyring.Config{
		ServiceName:              config.AppName,
		KeychainTrustApplication: false,
		AllowedBackends:          backends,
		FileDir:                  keyringDir,
		FilePasswordFunc:         fileKeyringPasswordFunc(),
	}

	if shouldUseTimeout(runtime.GOOS, backend, dbusAddr) {
		return openKeyringWithTimeout(cfg, keyringOpenTimeout)
	}

	ring, err := keyringOpenFunc(cfg)
	if err != nil {
		return nil, fmt.Errorf("open keyring: %w", err)
	}

	return ring, nil
}

func normalizeBackend(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func allowedBackends(backend string) ([]keyring.BackendType, error) {
	switch backend {
	case "", "auto":
		return nil, nil
	case "keychain":
		return []keyring.BackendType{keyring.KeychainBackend}, nil
	case "file":
		return []keyring.BackendType{keyring.FileBackend}, nil
	default:
		return nil, fmt.Errorf("%w: %q", errInvalidBackend, backend)
	}
}

func shouldForceFileBackend(goos, backend, dbusAddr string) bool {
	return goos == "linux" && (backend == "" || backend == "auto") && dbusAddr == ""
}

func shouldUseTimeout(goos, backend, dbusAddr string) bool {
	return goos == "linux" && (backend == "" || backend == "auto") && dbusAddr != ""
}

func fileKeyringPasswordFunc() keyring.PromptFunc {
	password := os.Getenv(keyringPasswordEnv)
	if password != "" {
		return keyring.FixedStringPrompt(password)
	}

	if term.IsTerminal(int(os.Stdin.Fd())) {
		return keyring.TerminalPrompt
	}

	return func(_ string) (string, error) {
		return "", fmt.Errorf("%w; set %s", errNoTTY, keyringPasswordEnv)
	}
}

type keyringResult struct {
	ring keyring.Keyring
	err  error
}

func openKeyringWithTimeout(cfg keyring.Config, timeout time.Duration) (keyring.Keyring, error) {
	ch := make(chan keyringResult, 1)

	go func() {
		ring, err := keyringOpenFunc(cfg)
		ch <- keyringResult{ring, err}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return nil, fmt.Errorf("open keyring: %w", res.err)
		}

		return res.ring, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("%w after %v; set %s=file and %s=<password>",
			errKeyringTimeout, timeout, keyringBackendEnv, keyringPasswordEnv)
	}
}

func OpenDefault() (Store, error) {
	ring, err := openKeyringFunc()
	if err != nil {
		return nil, err
	}

	return &KeyringStore{ring: ring}, nil
}

func (s *KeyringStore) SetToken(tok Token) error {
	if tok.CreatedAt.IsZero() {
		tok.CreatedAt = time.Now().UTC()
	}

	payload, err := json.Marshal(tok)
	if err != nil {
		return fmt.Errorf("encode token: %w", err)
	}

	if err := s.ring.Set(keyring.Item{
		Key:  tokenKey,
		Data: payload,
	}); err != nil {
		return wrapKeychainError(fmt.Errorf("store token: %w", err))
	}

	s.tokenLoaded = true
	s.tokenCache = &tok

	return nil
}

func (s *KeyringStore) GetToken() (Token, error) {
	if s.tokenLoaded {
		if s.tokenCache == nil {
			return Token{}, ErrNoToken
		}

		return *s.tokenCache, nil
	}

	item, err := s.ring.Get(tokenKey)
	if err != nil {
		s.tokenLoaded = true
		s.tokenCache = nil

		if errors.Is(err, keyring.ErrKeyNotFound) {
			return Token{}, ErrNoToken
		}

		return Token{}, fmt.Errorf("read token: %w", err)
	}

	var tok Token
	if err := json.Unmarshal(item.Data, &tok); err != nil {
		return Token{}, fmt.Errorf("decode token: %w", err)
	}

	s.tokenLoaded = true
	s.tokenCache = &tok

	return tok, nil
}

func (s *KeyringStore) DeleteToken() error {
	if err := s.ring.Remove(tokenKey); err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
		return fmt.Errorf("delete token: %w", err)
	}

	s.tokenLoaded = true
	s.tokenCache = nil

	return nil
}

func (s *KeyringStore) SetCredentials(creds OAuthCredentials) error {
	if strings.TrimSpace(creds.ClientID) == "" {
		return errMissingClientID
	}

	if strings.TrimSpace(creds.ClientSecret) == "" {
		return errMissingClientSecret
	}

	payload, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("encode credentials: %w", err)
	}

	if err := s.ring.Set(keyring.Item{
		Key:  credentialsKey,
		Data: payload,
	}); err != nil {
		return wrapKeychainError(fmt.Errorf("store credentials: %w", err))
	}

	s.credsLoaded = true
	s.credsCache = &creds

	return nil
}

func (s *KeyringStore) GetCredentials() (OAuthCredentials, error) {
	if s.credsLoaded {
		if s.credsCache == nil {
			return OAuthCredentials{}, ErrNoCredentials
		}

		return *s.credsCache, nil
	}

	item, err := s.ring.Get(credentialsKey)
	if err != nil {
		s.credsLoaded = true
		s.credsCache = nil

		if errors.Is(err, keyring.ErrKeyNotFound) {
			return OAuthCredentials{}, ErrNoCredentials
		}

		return OAuthCredentials{}, fmt.Errorf("read credentials: %w", err)
	}

	var creds OAuthCredentials
	if err := json.Unmarshal(item.Data, &creds); err != nil {
		return OAuthCredentials{}, fmt.Errorf("decode credentials: %w", err)
	}

	s.credsLoaded = true
	s.credsCache = &creds

	return creds, nil
}

func (s *KeyringStore) DeleteCredentials() error {
	if err := s.ring.Remove(credentialsKey); err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
		return fmt.Errorf("delete credentials: %w", err)
	}

	s.credsLoaded = true
	s.credsCache = nil

	return nil
}

func (s *KeyringStore) CredentialsExists() (bool, error) {
	if s.credsLoaded {
		return s.credsCache != nil, nil
	}

	_, err := s.GetCredentials()
	if err != nil {
		if errors.Is(err, ErrNoCredentials) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func wrapKeychainError(err error) error {
	if err == nil {
		return nil
	}

	if IsKeychainLockedError(err.Error()) {
		return fmt.Errorf("%w\n\nYour macOS keychain is locked. Run:\n  security unlock-keychain ~/Library/Keychains/login.keychain-db", err)
	}

	return err
}

func IsKeychainLockedError(msg string) bool {
	return strings.Contains(msg, "keychain is locked") ||
		strings.Contains(msg, "The user name or passphrase you entered is not correct")
}
