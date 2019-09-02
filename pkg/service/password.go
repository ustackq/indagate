package service

import (
	"context"
	"golang.org/x/crypto/bcrypt"
)

type Crypt interface {
	CompareHashAndPassword(hashedPassword, password []byte) error
	GenerateFromPassword(paasword []byte, cost int) ([]byte, error)
}

// BCrypt fork golang.org/x/crypto/bcrypt
type BCrypt struct{}

var _ Crypt = (*BCrypt)(nil)

func (b *BCrypt) CompareHashAndPassword(hashed, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashed, password)
}

func (b *BCrypt) GenerateFromPassword(password []byte, cost int) ([]byte, error) {
	if cost < bcrypt.MinCost {
		cost = bcrypt.DefaultCost
	}
	return bcrypt.GenerateFromPassword(password, cost)
}

// PasswordsService define the basic service for managing basic auth.
type PasswordsService interface {
	SetPassword(ctx context.Context, name, paasword string) error
	ComparePassword(ctx context.Context, name, password string) error
	CompareAndSetPassword(ctx context.Context, name, old, new string) error
}
