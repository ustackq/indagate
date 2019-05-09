package store

import (
	"context"
	"fmt"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

var (
	authBucket = []byte("authorization")
	authIndex  = []byte("authorizationIndex")
)

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
