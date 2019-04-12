package service

import (
	"context"
)

// IsInstallingResponse define installing response
type IsInstallingResponse struct {
	Allowed bool `json:"allowed"`
}

// SetupService define setup service for the first running.
type SetupService interface {
	IsInstalling(ctx context.Context) (bool, error)
	Setup(ctx context.Context, r *SetupRequest) (*SetupResponse, error)
}

// SetupRequest define setup request
type SetupRequest struct {
}

// SetupResponse define first run result
type SetupResponse struct {
}
