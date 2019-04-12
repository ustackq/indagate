package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrAuthHeaderMissing = errors.New("authorization Header is missing")
	ErrAuthBadScheme     = errors.New("authorization Header Scheme is invalid")

	// Make this to be bear token
	tokenScheme = "Token"
)

// GetToken return token from header.
func GetToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", ErrAuthHeaderMissing
	}
	if !strings.HasPrefix(header, tokenScheme) {
		return "", ErrAuthBadScheme
	}
	return header[len(tokenScheme):], nil
}

// SetToken adds token to the request.
func SetToken(r *http.Request, token string) {
	r.Header.Set("Authorization", fmt.Sprintf("%s%s", tokenScheme, token))
}

func ProbeAuthScheme(r *http.Request) (string, error) {
	_, err := GetToken(r)
	if err != nil {
		return "", errors.New("token required")
	}
	return tokenAuthScheme, nil
}
