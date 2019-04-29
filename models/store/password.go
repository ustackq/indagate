package store

import (
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
