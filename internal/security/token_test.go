package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestJWTIssuerIssue(t *testing.T) {
	const secret = "0123456789abcdef0123456789abcdef"
	fixedNow := time.Date(2026, time.July, 10, 13, 0, 0, 0, time.UTC)

	issuer, err := NewJWTIssuer(secret, 24*time.Hour)
	if err != nil {
		t.Fatalf("NewJWTIssuer() error = %v", err)
	}
	issuer.now = func() time.Time { return fixedNow }

	token, expiresAt, err := issuer.Issue("user-123")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if !expiresAt.Equal(fixedNow.Add(24 * time.Hour)) {
		t.Fatalf("expiresAt = %v, want %v", expiresAt, fixedNow.Add(24*time.Hour))
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("token contains %d segments, want 3", len(parts))
	}

	var header jwtHeader
	decodeSegment(t, parts[0], &header)
	if header.Algorithm != "HS256" || header.Type != "JWT" {
		t.Fatalf("unexpected header: %+v", header)
	}

	var claims jwtClaims
	decodeSegment(t, parts[1], &claims)
	if claims.Subject != "user-123" {
		t.Fatalf("subject = %q, want user-123", claims.Subject)
	}
	if claims.IssuedAt != fixedNow.Unix() {
		t.Fatalf("issued at = %d, want %d", claims.IssuedAt, fixedNow.Unix())
	}
	if claims.ExpiresAt != fixedNow.Add(24*time.Hour).Unix() {
		t.Fatalf("expires at = %d, want %d", claims.ExpiresAt, fixedNow.Add(24*time.Hour).Unix())
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	wantSignature := mac.Sum(nil)
	gotSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("decode signature: %v", err)
	}
	if !hmac.Equal(gotSignature, wantSignature) {
		t.Fatal("JWT signature is invalid")
	}
}

func TestNewJWTIssuerValidation(t *testing.T) {
	_, err := NewJWTIssuer("too-short", time.Hour)
	if !errors.Is(err, ErrJWTSecretTooShort) {
		t.Fatalf("error = %v, want %v", err, ErrJWTSecretTooShort)
	}

	_, err = NewJWTIssuer("0123456789abcdef0123456789abcdef", 0)
	if !errors.Is(err, ErrJWTExpiryInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrJWTExpiryInvalid)
	}
}

func decodeSegment(t *testing.T, segment string, target any) {
	t.Helper()

	payload, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		t.Fatalf("decode JWT segment: %v", err)
	}
	if err := json.Unmarshal(payload, target); err != nil {
		t.Fatalf("unmarshal JWT segment: %v", err)
	}
}
