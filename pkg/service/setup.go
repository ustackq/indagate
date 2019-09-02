package service

import (
	"context"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

// IsInstallingResponse define installing response
type IsInstallingResponse struct {
	Allowed bool `json:"allowed"`
}

// SetupService define setup service for the first running.
type SetupService interface {
	PasswordsService
	BucketService
	OrganizationService
	UserService
	AuthorizationService
	IsInstalling(ctx context.Context) (bool, error)
	Setup(ctx context.Context, r *SetupRequest) (*SetupResult, error)
}

// SetupResult define setup request
type SetupResult struct {
	User   *User          `json:"user"`
	Org    *Organization  `json:"org"`
	Bucket *Bucket        `json:"bucket"`
	Auth   *Authorization `json:"auth"`
}

// SetupRequest define first run result
type SetupRequest struct {
	User            string `json:"user"`
	Password        string `json:"password"`
	Org             string `json:"org"`
	Bucket          string `json:"bucket"`
	RetentionPeriod uint   `json:"retentionPeriods,omitempty"`
	Token           string `json:"token,omitempty"`
}

func (s *SetupRequest) Valid() error {
	if s.Password == "" {
		return &errors.Error{
			Code: errors.EmptyValue,
			Msg:  "password is empty",
		}
	}

	if s.User == "" {
		return &errors.Error{
			Code: errors.EmptyValue,
			Msg:  "user is empty",
		}
	}

	if s.Org == "" {
		return &errors.Error{
			Code: errors.EmptyValue,
			Msg:  "org is empty",
		}
	}

	if s.Bucket == "" {
		return &errors.Error{
			Code: errors.EmptyValue,
			Msg:  "bucket name is empty",
		}
	}

	return nil
}
