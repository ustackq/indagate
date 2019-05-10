package store

import (
	"context"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

var (
	authBucket = []byte("authorization")
	authIndex  = []byte("authorizationIndex")
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// assert Service implement service.AuthorizationService
//var _ service.AuthorizationService = (*Service)(nil)

func (s *Service) initializeAuth(ctx context.Context, tx Impl) error {
	if _, err := tx.Bucket(authBucket); err != nil {
		return err
	}
	if _, err := authIndexBucket(tx); err != nil {
		return err
	}
	return nil
}

func authIndexBucket(tx Impl) (Bucket, error) {
	b, err := tx.Bucket([]byte(authIndex))
	if err != nil {
		return nil, UnexpectedAuthIndexError(err)
	}
	return b, nil
}

func UnexpectedAuthIndexError(err error) *errors.Error {
	return &errors.Error{
		Code: errors.Internal,
		Msg:  fmt.Sprintf("unexpected error retrieving auth index, %v", err),
		Op:   "authIndex",
	}
}

// implememnt AuthorizationService

func (s *Service) FindAuthorizationByID(ctx context.Context, id service.ID) (*service.Authorization, error) {
	var a *service.Authorization
	err := s.store.View(ctx, func(tx Impl) error {
		//
		return nil
	})
	return a, err
}

func (s *Service) findAuthorizationByID(ctx context.Context, tx Impl, id service.ID) (*service.Authorization, error) {
	encodeID, err := id.Encode()
	if err != nil {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Err:  err,
		}
	}

	b, err := tx.Bucket(authBucket)
	if err != nil {
		return nil, err
	}

	v, err := b.Get(encodeID)
	if errors.IsNotFound(err) {
		return nil, &errors.Error{
			Code: errors.NotFound,
			Msg:  "authorization not found",
		}
	}

	if err != nil {
		return nil, err
	}

	auth := &service.Authorization{}
	if err := decodeAuthorization(v, auth); err != nil {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Err:  err,
		}
	}

	return auth, nil
}

func decodeAuthorization(v []byte, auth *service.Authorization) error {
	if err := json.Unmarshal(v, auth); err != nil {
		return nil
	}

	if auth.Active == "" {
		auth.Active = service.Active
	}

	return nil
}

func (s *Service) FindAuthorizationByToken(ctx context.Context, str string) (*service.Authorization, error) {
	var auth *service.Authorization

	return auth, nil
}

// TODO
func (s *Service) findAuthorizationByToken(ctx context.Context, tx Impl, str string) (*service.Authorization, error) {

}
