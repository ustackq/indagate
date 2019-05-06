package generator

import (
	"crypto/rand"
	"encoding/base64"
)

// TokenGenerator define token generator
type TokenGenerator interface {
	Token() (string, error)
}

type tokenGenerator struct {
	size int
}

func NewTokenGenerator(size int) TokenGenerator {
	return &tokenGenerator{
		size: size,
	}
}

func (tg *tokenGenerator) Token() (string, error) {
	b, err := tg.generateRandomBytes(tg.size)
	return base64.URLEncoding.EncodeToString(b), err
}

func (tg *tokenGenerator) generateRandomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
