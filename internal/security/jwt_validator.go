package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrTokenInvalid = errors.New("token is invalid")
	ErrTokenExpired = errors.New("token has expired")
)

// TokenClaims contains the authenticated identity extracted from a JWT.
type TokenClaims struct {
	Subject   string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// TokenValidator validates a signed access token and returns its claims.
type TokenValidator interface {
	Validate(token string) (*TokenClaims, error)
}

// JWTValidator validates HS256 JWT access tokens.
type JWTValidator struct {
	secret []byte
	now    func() time.Time
}

func NewJWTValidator(secret string) (*JWTValidator, error) {
	secret = strings.TrimSpace(secret)
	if len(secret) < minimumJWTSecretLength {
		return nil, ErrJWTSecretTooShort
	}

	return &JWTValidator{
		secret: []byte(secret),
		now:    func() time.Time { return time.Now().UTC() },
	}, nil
}

func (v *JWTValidator) Validate(token string) (*TokenClaims, error) {
	segments := strings.Split(token, ".")
	if len(segments) != 3 {
		return nil, ErrTokenInvalid
	}

	var header jwtHeader
	if err := decodeJWTSegment(segments[0], &header); err != nil {
		return nil, ErrTokenInvalid
	}
	if header.Algorithm != "HS256" || header.Type != "JWT" {
		return nil, ErrTokenInvalid
	}

	signature, err := base64.RawURLEncoding.DecodeString(segments[2])
	if err != nil {
		return nil, ErrTokenInvalid
	}

	signingInput := segments[0] + "." + segments[1]
	mac := hmac.New(sha256.New, v.secret)
	_, _ = mac.Write([]byte(signingInput))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return nil, ErrTokenInvalid
	}

	var claims jwtClaims
	if err := decodeJWTSegment(segments[1], &claims); err != nil {
		return nil, ErrTokenInvalid
	}
	if claims.Subject == "" || claims.IssuedAt <= 0 || claims.ExpiresAt <= 0 {
		return nil, ErrTokenInvalid
	}

	now := v.now().UTC()
	expiresAt := time.Unix(claims.ExpiresAt, 0).UTC()
	if !expiresAt.After(now) {
		return nil, ErrTokenExpired
	}

	return &TokenClaims{
		Subject:   claims.Subject,
		IssuedAt:  time.Unix(claims.IssuedAt, 0).UTC(),
		ExpiresAt: expiresAt,
	}, nil
}

func decodeJWTSegment(segment string, destination any) error {
	payload, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return err
	}

	return json.Unmarshal(payload, destination)
}
