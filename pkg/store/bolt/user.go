package bolt

import (
	"context"
	"encoding/json"

	bolt "go.etcd.io/bbolt"

	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

var (
	userBucket         = []byte("usersv1")
	userIndex          = []byte("userindexv1")
	userpasswordBucket = []byte("userpasswordv1")
)

func userIndexKey(key string) []byte {
	return []byte(key)
}

func (c *Client) findUserByID(ctx context.Context, tx *bolt.Tx, id service.ID) (*service.User, *errors.Error) {
	encodedID, err := id.Encode()
	if err != nil {
		return nil, &errors.Error{
			Err: err,
		}
	}

	var user service.User
	v := tx.Bucket(userBucket).Get(encodedID)

	if len(v) == 0 {
		return nil, &errors.Error{
			Code: errors.NotFound,
			Msg:  "user not found",
		}
	}

	if err := json.Unmarshal(v, &user); err != nil {
		return nil, &errors.Error{
			Err: err,
		}
	}
	return &user, nil
}

func (c *Client) findUserByName(ctx context.Context, tx *bolt.Tx, name string) (*service.User, *errors.Error) {
	user := tx.Bucket(userIndex).Get(userIndexKey(name))
	if user == nil {
		return nil, &errors.Error{
			Code: errors.NotFound,
			Msg:  "user not found",
		}
	}

	var id service.ID
	if err := id.Decode(user); err != nil {
		return nil, &errors.Error{
			Err: err,
		}
	}

	return c.findUserByID(ctx, tx, id)
}
