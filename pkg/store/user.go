package store

import (
	"context"
	"fmt"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

var (
	userBucket = []byte("userv1alpha1")
	userIndex  = []byte("userIndexv1alpha1")
)

func (s *Service) initializaUsers(ctx context.Context, tx Impl) error {
	if _, err := s.userBucket(tx); err != nil {
		return err
	}

	if _, err := s.userIndexBucket(tx); err != nil {
		return err
	}
	return nil
}

func (s *Service) userBucket(tx Impl) (Bucket, error) {
	b, err := tx.Bucket([]byte(userBucket))
	if err != nil {
		return nil, UnexpectedUserError(err)
	}
	return b, nil
}

func (s *Service) userIndexBucket(tx Impl) (Bucket, error) {
	b, err := tx.Bucket([]byte(userIndex))
	if err != nil {
		return nil, UnexpectedUserIndexError(err)
	}

	return b, nil
}

func UnexpectedUserError(err error) *errors.Error {
	return &errors.Error{
		Code: errors.Internal,
		Msg:  fmt.Sprintf("unexpected error retrieving user bucket; %v", err),
		Op:   "userBucket",
	}
}

func UnexpectedUserIndexError(err error) *errors.Error {
	return &errors.Error{
		Code: errors.Internal,
		Msg:  fmt.Sprintf("unexpected error retrieving user index; %v", err),
		Op:   "userBucket",
	}
}
