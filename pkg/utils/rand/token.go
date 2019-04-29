package rand

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/ustackq/indagate/pkg/service"
)

type TokenGenerator struct {
	size int
}

func NewTokenGenerator(n int) service.TokenGenerator {
	return &TokenGenerator{
		size: n,
	}
}

func (t *TokenGenerator) Token() (string, error) {
	b, err := generateRandomString(t.size)
	return b, err
}

func generateRandomString(n int) (string, error) {
	b, err := generateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}

// TODO: rand need seed
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
