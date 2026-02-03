package errfmt

import (
	"errors"

	"github.com/dedene/raindrop-cli/internal/api"
	"github.com/dedene/raindrop-cli/internal/auth"
)

// Sentinel errors for user-friendly messages.
var (
	errAuthFailed         = errors.New("authentication failed\n  Run: raindrop auth token <token>\n  or:  raindrop auth login")
	errNotAuthenticated   = errors.New("not authenticated\n  Run: raindrop auth token <token>\n  or:  raindrop auth login")
	errRateLimitExceeded  = errors.New("rate limit exceeded\n  Wait and try again")
	errAuthExpired        = errors.New("authentication expired\n  Run: raindrop auth token <token>\n  or:  raindrop auth login")
	errAccessDenied       = errors.New("access denied\n  Check your token permissions")
	errNotFound           = errors.New("not found")
	errCredentialsMissing = errors.New("oauth client not configured\n  Run: raindrop auth setup <client_id>")
)

// ExitError wraps an error with an exit code.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}

	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func (e *ExitError) ExitCode() int {
	if e == nil {
		return 1
	}

	return e.Code
}

// Format wraps an error with user-friendly hints and exit codes.
func Format(err error) error {
	if err == nil {
		return nil
	}

	// Check for auth errors
	var authErr *api.AuthError
	if errors.As(err, &authErr) {
		return &ExitError{
			Code: api.ExitAuth,
			Err:  errAuthFailed,
		}
	}

	if errors.Is(err, auth.ErrNotAuthenticated) {
		return &ExitError{
			Code: api.ExitAuth,
			Err:  errNotAuthenticated,
		}
	}

	if errors.Is(err, auth.ErrNoCredentials) {
		return &ExitError{
			Code: api.ExitAuth,
			Err:  errCredentialsMissing,
		}
	}

	// Check for API errors
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		return &ExitError{
			Code: apiErr.ExitCode(),
			Err:  formatAPIError(apiErr),
		}
	}

	// Check for rate limit errors
	var rlErr *api.RateLimitError
	if errors.As(err, &rlErr) {
		return &ExitError{
			Code: api.ExitRateLimit,
			Err:  errRateLimitExceeded,
		}
	}

	// Check for not found errors
	var nfErr *api.NotFoundError
	if errors.As(err, &nfErr) {
		return &ExitError{
			Code: api.ExitNotFound,
			Err:  err,
		}
	}

	return err
}

func formatAPIError(err *api.APIError) error {
	switch err.StatusCode {
	case 401:
		return errAuthExpired
	case 403:
		return errAccessDenied
	case 404:
		return errNotFound
	case 429:
		return errRateLimitExceeded
	default:
		return err
	}
}
