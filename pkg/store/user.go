package store

import (
	"context"
	"fmt"
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

var (
	userBucket = []byte("userv1alpha1")
	userIndex  = []byte("userIndexv1alpha1")

	ErrUserNotFound = &errors.Error{
		Msg:  "user not found",
		Code: errors.NotFound,
	}
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

func (s *Service) findUserByName(ctx context.Context, tx Impl, str string) (*service.User, error) {
	b, err := s.userIndexBucket(tx)
	if err != nil {
		return nil, err
	}

	uid, err := b.Get([]byte(str))
	if err == ErrKeyNotFound {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, &errors.Error{
			Code: errors.Internal,
			Err:  err,
		}
	}

	var id service.ID
	if err := id.Decode(uid); err != nil {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Err:  err,
		}
	}

	return s.findUserByID(ctx, tx, id)
}

func (s *Service) findUserByID(ctx context.Context, tx Impl, id service.ID) (*service.User, error) {
	encodedID, err := id.Encode()
	if err != nil {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Err:  err,
		}
	}

	b, err := s.userBucket(tx)
	if err != nil {
		return nil, err
	}

	v, err := b.Get(encodedID)
	if errors.IsNotFound(err) {
		return nil, &errors.Error{
			Code: errors.NotFound,
			Msg:  "user not found",
		}
	}

	if err != nil {
		return nil, &errors.Error{
			Code: errors.Internal,
			Err:  err,
		}
	}
	return unmarshalUser(v)
}

func unmarshalUser(v []byte) (*service.User, error) {
	u := &service.User{}
	if err := json.Unmarshal(v, u); err != nil {
		return nil, &errors.Error{
			Code: errors.Internal,
			Msg:  "user could not be unmarshalled",
			Err:  err,
			Op:   "unmarshalUser",
		}
	}
	return u, nil
}
