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
var _ service.AuthorizationService = (*Service)(nil)

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
	err := s.store.View(ctx, func(tx Impl) error {
		a, err := s.findAuthorizationByToken(ctx, tx, str)
		if err != nil {
			return err
		}
		auth = a
		return nil
	})
	if err != nil {
		return nil, err
	}
	return auth, nil
}

func (s *Service) findAuthorizationByToken(ctx context.Context, tx Impl, str string) (*service.Authorization, error) {
	idx, err := authIndexBucket(tx)
	if err != nil {
		return nil, err
	}

	auth, err := idx.Get([]byte(str))
	if errors.IsNotFound(err) {
		return nil, &errors.Error{
			Code: errors.NotFound,
			Msg:  "authorization not found",
		}
	}

	var id service.ID
	if err := id.Decode(auth); err != nil {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Err:  err,
		}
	}
	return s.findAuthorizationByID(ctx, tx, id)
}

// FindAuthorization list Authorization filter by arg
func (s *Service) FindAuthorization(ctx context.Context, filter service.AuthorizationFilter, opt ...service.FindOptions) ([]*service.Authorization, int, error) {
	if filter.Token != nil {
		auth, err := s.FindAuthorizationByToken(ctx, *filter.Token)
		if err != nil {
			return nil, 0, &errors.Error{
				Err: err,
			}
		}
		return []*service.Authorization{auth}, 1, nil
	}

	if filter.ID != nil {
		auth, err := s.FindAuthorizationByID(ctx, *filter.ID)
		if err != nil {
			return nil, 0, &errors.Error{
				Err: err,
			}
		}
		return []*service.Authorization{auth}, 1, nil
	}

	auths := []*service.Authorization{}
	err := s.store.View(ctx, func(tx Impl) error {
		a, err := s.findAuthorization(ctx, tx, filter)
		if err != nil {
			return err
		}
		auths = a
		return nil
	})

	if err != nil {
		return nil, 0, &errors.Error{
			Err: err,
		}
	}

	return auths, len(auths), nil
}

func filterAuthorizationFn(filter service.AuthorizationFilter) func(auth *service.Authorization) bool {
	if filter.ID != nil {
		return func(auth *service.Authorization) bool {
			return auth.ID == *filter.ID
		}
	}

	if filter.Token != nil {
		return func(auth *service.Authorization) bool {
			return auth.Token == *filter.Token
		}
	}

	// compare org and user
	if filter.OrgID != nil && filter.UserID != nil {
		return func(auth *service.Authorization) bool {
			return auth.OrgID == *filter.OrgID && auth.UserID == *filter.UserID
		}
	}

	if filter.OrgID != nil {
		return func(auth *service.Authorization) bool {
			return auth.OrgID == *filter.OrgID
		}
	}

	if filter.UserID != nil {
		return func(auth *service.Authorization) bool {
			return auth.UserID == *filter.UserID
		}
	}

	return func(auth *service.Authorization) bool { return true }
}

func (s *Service) findAuthorization(ctx context.Context, tx Impl, f service.AuthorizationFilter) ([]*service.Authorization, error) {
	if f.User != nil {
		u, err := s.findUserByName(ctx, tx, *f.User)
		if err != nil {
			return nil, err
		}
		f.UserID = &u.ID
	}

	if f.Org != nil {
		org, err := s.findOrgnizationByName(ctx, tx, *f.Org)
		if err != nil {
			return nil, err
		}
		f.OrgID = &org.ID
	}

	auths := []*service.Authorization{}
	filterFn := filterAuthorizationFn(f)
	err := s.forEachAuthorization(ctx, tx, func(auth *service.Authorization) bool {
		if filterFn(auth) {
			auths = append(auths, auth)
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return auths, nil
}

func (s *Service) forEachAuthorization(ctx context.Context, tx Impl, fn func(*service.Authorization) bool) error {
	b, err := tx.Bucket(authBucket)
	if err != nil {
		return err
	}

	cur, err := b.Cursor()
	if err != nil {
		return err
	}

	for k, v := cur.First(); k != nil; k, v = cur.Next() {
		auth := &service.Authorization{}
		if err := json.Unmarshal(v, auth); err != nil {
			return err
		}
		if auth.Active == "" {
			auth.Active = service.Active
		}
		if !fn(auth) {
			break
		}
	}
	return nil
}

func (s *Service) CreateAuthorization(ctx context.Context, auth *service.Authorization) error {
	return s.store.Modify(ctx, func(tx Impl) error {
		return s.createAuthorization(ctx, tx, auth)
	})
}

func (s *Service) uniqueAuthToken(ctx context.Context, tx Impl, auth *service.Authorization) error {
	err := s.unique(ctx, tx, authIndex, []byte(auth.Token))
	if err == NotUniqueError {
		return service.ErrCreateToken
	}
	return err
}

func (s *Service) createAuthorization(ctx context.Context, tx Impl, auth *service.Authorization) error {
	if err := auth.Valid(); err != nil {
		return err
	}
	if _, err := s.findUserByID(ctx, tx, auth.UserID); err != nil {
		return err
	}
	if _, err := s.findOrgnizationByID(ctx, tx, auth.OrgID); err != nil {
		return err
	}

	if err := s.uniqueAuthToken(ctx, tx, auth); err != nil {
		return err
	}

	if auth.Token == "" {
		token, err := s.TokenGenerator.Token()
		if err != nil {
			return &errors.Error{
				Err: err,
			}
		}
		auth.Token = token
	}
	auth.ID = s.IDGenerator.ID()
	if err := s.putAuthorization(ctx, tx, auth); err != nil {
		return err
	}

	return nil
}

func encodeAuth(auth *service.Authorization) ([]byte, error) {
	switch auth.Active {
	case "":
		auth.Active = service.Active
	case service.Active, service.Inactive:
	default:
		return nil, &errors.Error{
			Code: errors.Invalid,
			Msg:  "unknown auth status",
		}
	}
	return json.Marshal(auth)
}

func (s *Service) putAuthorization(ctx context.Context, tx Impl, auth *service.Authorization) error {
	v, err := encodeAuth(auth)
	if err != nil {
		// TODO: using fn wrapper
		return &errors.Error{
			Code: errors.Invalid,
			Err:  err,
		}
	}

	// encode id
	encodeID, err := auth.ID.Encode()
	if err != nil {
		return errors.NotFoundErr(err)
	}

	idx, err := authIndexBucket(tx)
	if err != nil {
		return err
	}

	if err := idx.Put([]byte(auth.Token), encodeID); err != nil {
		return errors.InternalErr(err)
	}

	b, err := tx.Bucket(authBucket)
	if err != nil {
		return err
	}

	if err := b.Put(encodeID, v); err != nil {
		return errors.WrapperErr(err)
	}

	return nil
}

func (s *Service) DeleteAuthorization(ctx context.Context, id service.ID) error {
	return s.store.Modify(ctx, func(tx Impl) error {
		return s.deleteAuthorization(ctx, tx, id)
	})
}

func (s *Service) deleteAuthorization(ctx context.Context, tx Impl, id service.ID) error {
	auth, err := s.findAuthorizationByID(ctx, tx, id)
	if err != nil {
		return err
	}

	idx, err := authIndexBucket(tx)
	if err != nil {
		return err
	}

	if err := idx.Delete([]byte(auth.Token)); err != nil {
		return errors.WrapperErr(err)
	}

	encodeID, err := id.Encode()
	if err != nil {
		return errors.WrapperErr(err)
	}

	b, err := tx.Bucket(authBucket)
	if err != nil {
		return err
	}

	if err := b.Delete(encodeID); err != nil {
		return errors.WrapperErr(err)
	}
	return nil
}

func (s *Service) UpdateAuthorization(ctx context.Context, id service.ID, update *service.AuthorizationUpdate) error {
	return s.store.Modify(ctx, func(tx Impl) error {
		return s.updateAuthorization(ctx, tx, id, update)
	})
}

func (s *Service) updateAuthorization(ctx context.Context, tx Impl, id service.ID, update *service.AuthorizationUpdate) error {
	auth, err := s.findAuthorizationByID(ctx, tx, id)
	if err != nil {
		return err
	}
	if update.Status != nil {
		auth.Active = *update.Status
	}
	if update.Description != nil {
		auth.Description = *update.Description
	}

	v, err := encodeAuth(auth)
	if err != nil {
		return err
	}

	encodeID, err := id.Encode()
	if err != nil {
		return errors.WrapperErr(err)
	}

	b, err := tx.Bucket(authBucket)
	if err != nil {
		return err
	}

	if err := b.Put(encodeID, v); err != nil {
		return errors.WrapperErr(err)
	}
	return nil
}
