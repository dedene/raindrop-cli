package api

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

const (
	MaxRateLimitRetries   = 3
	Max5xxRetries         = 1
	RateLimitBaseDelay    = 1 * time.Second
	ServerErrorRetryDelay = 2 * time.Second
)

// RetryTransport wraps an http.RoundTripper with retry logic for
// rate limits (429) and server errors (5xx).
type RetryTransport struct {
	Base          http.RoundTripper
	MaxRetries429 int
	MaxRetries5xx int
	BaseDelay     time.Duration
}

func NewRetryTransport(base http.RoundTripper) *RetryTransport {
	if base == nil {
		base = http.DefaultTransport
	}

	return &RetryTransport{
		Base:          base,
		MaxRetries429: MaxRateLimitRetries,
		MaxRetries5xx: Max5xxRetries,
		BaseDelay:     RateLimitBaseDelay,
	}
}

func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := ensureReplayableBody(req); err != nil {
		return nil, err
	}

	var resp *http.Response
	var err error
	retries429 := 0
	retries5xx := 0

	for {
		if req.GetBody != nil {
			if req.Body != nil {
				_ = req.Body.Close()
			}

			body, getErr := req.GetBody()
			if getErr != nil {
				return nil, fmt.Errorf("reset request body: %w", getErr)
			}

			req.Body = body
		}

		resp, err = t.Base.RoundTrip(req)
		if err != nil {
			return nil, fmt.Errorf("round trip: %w", err)
		}

		if resp.StatusCode < 400 {
			return resp, nil
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			if retries429 >= t.MaxRetries429 {
				return resp, nil
			}

			delay := t.calculateBackoff(retries429, resp)
			drainAndClose(resp.Body)

			if err := t.sleep(req.Context(), delay); err != nil {
				return nil, err
			}

			retries429++

			continue
		}

		if resp.StatusCode >= 500 {
			if retries5xx >= t.MaxRetries5xx {
				return resp, nil
			}

			drainAndClose(resp.Body)

			if err := t.sleep(req.Context(), ServerErrorRetryDelay); err != nil {
				return nil, err
			}

			retries5xx++

			continue
		}

		return resp, nil
	}
}

func (t *RetryTransport) calculateBackoff(attempt int, resp *http.Response) time.Duration {
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			if seconds < 0 {
				return 0
			}

			return time.Duration(seconds) * time.Second
		}

		if parsed, err := http.ParseTime(retryAfter); err == nil {
			d := time.Until(parsed)
			if d < 0 {
				return 0
			}

			return d
		}
	}

	if t.BaseDelay <= 0 {
		return 0
	}

	baseDelay := t.BaseDelay * time.Duration(1<<attempt)
	if baseDelay <= 0 {
		return 0
	}

	jitterRange := baseDelay / 2
	if jitterRange <= 0 {
		return baseDelay
	}

	jitter := time.Duration(rand.Int64N(int64(jitterRange))) //nolint:gosec // non-crypto jitter

	return baseDelay + jitter
}

func (t *RetryTransport) sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("sleep interrupted: %w", ctx.Err())
	}
}

type bytesReader struct {
	data []byte
	pos  int
}

func newBytesReader(data []byte) *bytesReader {
	return &bytesReader{data: data}
}

func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	n = copy(p, r.data[r.pos:])
	r.pos += n

	return n, nil
}

func ensureReplayableBody(req *http.Request) error {
	if req == nil || req.Body == nil || req.GetBody != nil {
		return nil
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("read request body: %w", err)
	}

	_ = req.Body.Close()

	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(newBytesReader(bodyBytes)), nil
	}
	req.Body = io.NopCloser(newBytesReader(bodyBytes))

	return nil
}

func drainAndClose(body io.ReadCloser) {
	if body == nil {
		return
	}

	_, _ = io.Copy(io.Discard, io.LimitReader(body, 1<<20))
	_ = body.Close()
}
