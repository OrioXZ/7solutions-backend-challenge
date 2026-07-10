package security

import (
	"errors"
	"strings"
	"testing"
	"time"
)

const testJWTSecret = "test-secret-at-least-thirty-two-characters-long"

func TestJWTValidatorAcceptsValidToken(t *testing.T) {
	issuedAt := time.Date(2026, time.July, 10, 12, 0, 0, 0, time.UTC)
	issuer, err := NewJWTIssuer(testJWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("NewJWTIssuer() error = %v", err)
	}
	issuer.now = func() time.Time { return issuedAt }

	token, _, err := issuer.Issue("user-123")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	validator, err := NewJWTValidator(testJWTSecret)
	if err != nil {
		t.Fatalf("NewJWTValidator() error = %v", err)
	}
	validator.now = func() time.Time { return issuedAt.Add(30 * time.Minute) }

	claims, err := validator.Validate(token)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if claims.Subject != "user-123" {
		t.Fatalf("Subject = %q, want %q", claims.Subject, "user-123")
	}
	if !claims.IssuedAt.Equal(issuedAt) {
		t.Fatalf("IssuedAt = %v, want %v", claims.IssuedAt, issuedAt)
	}
	if !claims.ExpiresAt.Equal(issuedAt.Add(time.Hour)) {
		t.Fatalf("ExpiresAt = %v, want %v", claims.ExpiresAt, issuedAt.Add(time.Hour))
	}
}

func TestJWTValidatorRejectsTamperedToken(t *testing.T) {
	issuer, err := NewJWTIssuer(testJWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("NewJWTIssuer() error = %v", err)
	}

	token, _, err := issuer.Issue("user-123")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	parts := strings.Split(token, ".")
	parts[1] = parts[1] + "x"
	tamperedToken := strings.Join(parts, ".")

	validator, err := NewJWTValidator(testJWTSecret)
	if err != nil {
		t.Fatalf("NewJWTValidator() error = %v", err)
	}

	if _, err := validator.Validate(tamperedToken); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrTokenInvalid)
	}
}

func TestJWTValidatorRejectsExpiredToken(t *testing.T) {
	issuedAt := time.Date(2026, time.July, 10, 12, 0, 0, 0, time.UTC)
	issuer, err := NewJWTIssuer(testJWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("NewJWTIssuer() error = %v", err)
	}
	issuer.now = func() time.Time { return issuedAt }

	token, _, err := issuer.Issue("user-123")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	validator, err := NewJWTValidator(testJWTSecret)
	if err != nil {
		t.Fatalf("NewJWTValidator() error = %v", err)
	}
	validator.now = func() time.Time { return issuedAt.Add(2 * time.Hour) }

	if _, err := validator.Validate(token); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrTokenExpired)
	}
}

func TestNewJWTValidatorRejectsShortSecret(t *testing.T) {
	if _, err := NewJWTValidator("too-short"); !errors.Is(err, ErrJWTSecretTooShort) {
		t.Fatalf("NewJWTValidator() error = %v, want %v", err, ErrJWTSecretTooShort)
	}
}
