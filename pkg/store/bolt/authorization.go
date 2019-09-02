package bolt

import (
	"context"
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"

	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

var (
	authorizationBucket = []byte("authorizationsv1")
	authorizationIndex  = []byte("authorizationindexv1")
)

var _ service.AuthorizationService = (*Client)(nil)

func (c *Client) initializaAuthorizations(ctx context.Context, tx *bolt.Tx) error {
	if _, err := tx.CreateBucketIfNotExists(authorizationBucket); err != nil {
		return nil
	}

	if _, err := tx.CreateBucketIfNotExists(authorizationIndex); err != nil {
		return nil
	}
	return nil
}

// FindAuthorizationByID return a service.Authorization instance by ID
func (c *Client) FindAuthorizationByID(ctx context.Context, id service.ID) (*service.Authorization, error) {
	var a *service.Authorization
	var err error
	err = c.db.View(func(tx *bolt.Tx) error {
		var se *errors.Error
		a, se := c.findAuthorizationByID(ctx, tx, id)
		if se != nil {
			se.Op = getOp(service.OpFindAuthorizationByID)
			err = se
		}
		return err
	})

	return a, err
}

func (c *Client) findAuthorizationByID(ctx context.Context, tx *bolt.Tx, id service.ID) (*service.Authorization, *errors.Error) {
	encodedID, err := id.Encode()
	if err != nil {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Err:  err,
		}
	}

	var a service.Authorization
	v := tx.Bucket(authorizationBucket).Get(encodedID)
	if len(v) <= 0 {
		return nil, &errors.Error{
			Code: errors.NotFound,
			Msg:  fmt.Sprintf("authorization %s not found", encodedID),
		}
	}

	if err := decodeAuthorization(v, &a); err != nil {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Err:  err,
		}
	}

	return &a, nil
}

func authorizationIndexKey(key string) []byte {
	return []byte(key)
}

// FindAuthorizationByToken return a service.Authorization instance by token.
func (c *Client) FindAuthorizationByToken(ctx context.Context, token string) (*service.Authorization, error) {
	var a *service.Authorization
	var err error
	err = c.db.View(func(tx *bolt.Tx) error {
		var se *errors.Error
		a, se = c.findAuthorizationByToken(ctx, tx, token)
		if se != nil {
			se.Op = getOp(service.OpFindAuthorizationByToken)
			err = se
		}
		return err
	})
	return a, err
}

func (c *Client) findAuthorizationByToken(ctx context.Context, tx *bolt.Tx, token string) (*service.Authorization, *errors.Error) {
	a := tx.Bucket(authorizationBucket).Get(authorizationIndexKey(token))
	if a == nil {
		return nil, &errors.Error{
			Code: errors.NotFound,
			Msg:  "authorization not found",
		}
	}

	var id service.ID
	if err := id.Decode(a); err != nil {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Err:  err,
		}
	}

	return c.findAuthorizationByID(ctx, tx, id)
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

	// Filter by orgnization and user
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

	return func(auth *service.Authorization) bool { return false }
}

// FindAuthorizations retives all authorizations that match a authorization filter.
// Filters using ID, or Token should be efficient.
func (c *Client) FindAuthorizations(ctx context.Context, filter service.AuthorizationFilter, opt ...service.FindOptions) ([]*service.Authorization, int, error) {
	if filter.ID != nil {
		auth, err := c.FindAuthorizationByID(ctx, *filter.ID)
		if err != nil {
			return nil, 0, &errors.Error{
				Err: err,
				Op:  getOp(service.OpFindAuthorizations),
			}
		}
		return []*service.Authorization{auth}, 1, nil
	}

	if filter.Token != nil {
		auth, err := c.FindAuthorizationByToken(ctx, *filter.Token)
		if err != nil {
			return nil, 0, &errors.Error{
				Err: err,
				Op:  getOp(service.OpFindAuthorizations),
			}
		}
		return []*service.Authorization{auth}, 1, nil
	}

	var auths []*service.Authorization
	err := c.db.View(func(tx *bolt.Tx) error {
		as, err := c.findAuthorizations(ctx, tx, filter)
		if err != nil {
			return err
		}
		auths = append(auths, as...)
		return nil
	})

	return auths, len(auths), err
}

func (c *Client) findAuthorizations(ctx context.Context, tx *bolt.Tx, filter service.AuthorizationFilter) ([]*service.Authorization, error) {
	// If User was provided, look up user by userID
	if filter.User != nil && filter.UserID == nil {
		user, err := c.findUserByName(ctx, tx, *filter.User)
		if err != nil {
			return nil, err
		}

		filter.UserID = &user.ID
	}

	if filter.Org != nil && filter.OrgID == nil {
		org, err := c.findOrganizationByName(ctx, tx, *filter.Org)
		if err != nil {
			return nil, err
		}
		filter.OrgID = &org.ID
	}

	var auths []*service.Authorization
	filterFn := filterAuthorizationFn(filter)
	err := c.forEachAuthorization(ctx, tx, func(auth *service.Authorization) bool {
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

// CreateAuthorization creates a service authorization and set auth.ID and auth.UserID if not provide.
func (c *Client) CreateAuthorization(ctx context.Context, auth *service.Authorization) error {
	op := getOp(service.OpCreateAuthorization)
	if err := auth.Valid(); err != nil {
		return &errors.Error{
			Err: err,
			Op:  op,
		}
	}

	return c.db.Update(func(tx *bolt.Tx) error {
		_, err := c.findUserByID(ctx, tx, auth.UserID)
		if err != nil {
			return service.ErrCreateToken
		}

		if unique := c.uniqueAuthorizationToken(ctx, tx, auth); !unique {
			return service.ErrCreateToken
		}

		// detect token wether exist or not.
		if auth.Token == "" {
			token, err := c.TokenGenerator.Token()
			if err != nil {
				return &errors.Error{
					Err: err,
					Op:  op,
				}
			}
			auth.Token = token
		}

		auth.ID = c.IDGenerator.ID()
		err = c.putAuthorization(ctx, tx, auth)
		if err != nil {
			err.Op = op
			return err
		}

		return nil
	})
}

func (c *Client) forEachAuthorization(ctx context.Context, tx *bolt.Tx, fn func(*service.Authorization) bool) error {
	cur := tx.Bucket(authorizationBucket).Cursor()
	for k, v := cur.First(); k != nil; k, v = cur.Next() {
		a := &service.Authorization{}

		if err := decodeAuthorization(v, a); err != nil {
			return nil
		}

		if !fn(a) {
			break
		}
	}

	return nil
}

func encodeAuthorization(auth *service.Authorization) ([]byte, error) {
	switch auth.Status {
	case service.Active, service.Inactive:
	case "":
		auth.Status = service.Active
	default:
		return nil, &errors.Error{
			Code: errors.Invalid,
			Msg:  "unknown authorization status",
		}
	}

	return json.Marshal(auth)
}

func decodeAuthorization(v []byte, auth *service.Authorization) error {
	if err := json.Unmarshal(v, auth); err != nil {
		return err
	}

	if auth.Status == "" {
		auth.Status = service.Active
	}
	return nil
}

func (c *Client) putAuthorization(ctx context.Context, tx *bolt.Tx, auth *service.Authorization) *errors.Error {
	v, err := encodeAuthorization(auth)
	if err != nil {
		return &errors.Error{
			Code: errors.Invalid,
			Err:  err,
		}
	}

	encodeID, err := auth.ID.Encode()
	if err != nil {
		return &errors.Error{
			Code: errors.NotFound,
			Err:  err,
		}
	}

	if err := tx.Bucket(authorizationBucket).Put(encodeID, v); err != nil {
		return &errors.Error{
			Err: err,
		}
	}

	return nil
}

func (c *Client) uniqueAuthorizationToken(ctx context.Context, tx *bolt.Tx, auth *service.Authorization) bool {
	return len(tx.Bucket(authorizationIndex).Get(authorizationIndexKey(auth.Token))) == 0
}

func (c *Client) deleteAuthorization(ctx context.Context, tx *bolt.Tx, id service.ID) *errors.Error {
	a, err := c.findAuthorizationByID(ctx, tx, id)
	if err != nil {
		return err
	}

	if err := tx.Bucket(authorizationIndex).Delete(authorizationIndexKey(a.Token)); err != nil {
		return &errors.Error{
			Err: err,
		}
	}
	return nil
}
