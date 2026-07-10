package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const minimumJWTSecretLength = 32

var (
	ErrJWTSecretTooShort = errors.New("JWT secret must be at least 32 characters")
	ErrJWTExpiryInvalid  = errors.New("JWT expiry must be greater than zero")
)

// TokenIssuer creates signed access tokens for authenticated users.
type TokenIssuer interface {
	Issue(subject string) (token string, expiresAt time.Time, err error)
}

type JWTIssuer struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type jwtClaims struct {
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func NewJWTIssuer(secret string, ttl time.Duration) (*JWTIssuer, error) {
	secret = strings.TrimSpace(secret)
	if len(secret) < minimumJWTSecretLength {
		return nil, ErrJWTSecretTooShort
	}
	if ttl <= 0 {
		return nil, ErrJWTExpiryInvalid
	}

	return &JWTIssuer{
		secret: []byte(secret),
		ttl:    ttl,
		now:    func() time.Time { return time.Now().UTC() },
	}, nil
}

func (i *JWTIssuer) Issue(subject string) (string, time.Time, error) {
	now := i.now().UTC()
	expiresAt := now.Add(i.ttl)

	headerSegment, err := encodeJWTSegment(jwtHeader{
		Algorithm: "HS256",
		Type:      "JWT",
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("encode JWT header: %w", err)
	}

	claimsSegment, err := encodeJWTSegment(jwtClaims{
		Subject:   subject,
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt.Unix(),
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("encode JWT claims: %w", err)
	}

	signingInput := headerSegment + "." + claimsSegment
	mac := hmac.New(sha256.New, i.secret)
	if _, err := mac.Write([]byte(signingInput)); err != nil {
		return "", time.Time{}, fmt.Errorf("sign JWT: %w", err)
	}

	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return signingInput + "." + signature, expiresAt, nil
}

func encodeJWTSegment(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(payload), nil
}
