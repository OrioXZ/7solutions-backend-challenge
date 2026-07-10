package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher abstracts password hashing so authentication services remain testable.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) error
}

type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{cost: bcrypt.DefaultCost}
}

func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return string(hash), nil
}

func (h *BcryptPasswordHasher) Verify(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("verify password: %w", err)
	}

	return nil
}
