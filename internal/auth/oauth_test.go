package auth

import (
	"errors"
	"testing"
)

func TestValidateOAuthState(t *testing.T) {
	t.Run("matching state", func(t *testing.T) {
		if err := validateOAuthState("abc123", "abc123"); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("missing state rejected", func(t *testing.T) {
		err := validateOAuthState("abc123", "")
		if !errors.Is(err, errStateMismatch) {
			t.Fatalf("expected errStateMismatch, got %v", err)
		}
	})

	t.Run("mismatched state rejected", func(t *testing.T) {
		err := validateOAuthState("abc123", "xyz789")
		if !errors.Is(err, errStateMismatch) {
			t.Fatalf("expected errStateMismatch, got %v", err)
		}
	})
}
