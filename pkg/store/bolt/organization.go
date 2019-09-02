package bolt

import (
	"context"
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"

	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/tracing"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

var (
	organizationBucket = []byte("organizationv1")
	organizationIndex  = []byte("organizationindexv1")
)

func organizationIndexKey(key string) []byte {
	return []byte(key)
}

func (c *Client) findOrganizationByName(ctx context.Context, tx *bolt.Tx, name string) (*service.Organization, error) {
	span, ctx := tracing.StartSpanFromContext(ctx)
	defer span.End()

	o := tx.Bucket(organizationIndex).Get(organizationIndexKey(name))
	if o == nil {
		return nil, &errors.Error{
			Code: errors.NotFound,
			Msg:  fmt.Sprintf("organization name '%s' not found", name),
		}
	}

	var id service.ID
	if err := id.Decode(o); err != nil {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Err:  err,
		}
	}

	return c.findOrganizationByID(ctx, tx, id)
}

func (c *Client) findOrganizationByID(ctx context.Context, tx *bolt.Tx, id service.ID) (*service.Organization, error) {
	span, _ := tracing.StartSpanFromContext(ctx)
	defer span.End()

	encodedID, err := id.Encode()
	if err != nil {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Err:  err,
		}
	}

	v := tx.Bucket(organizationBucket).Get(encodedID)
	if len(v) == 0 {
		return nil, &errors.Error{
			Code: errors.NotFound,
			Msg:  "organization not found",
		}
	}

	var org service.Organization
	if err := json.Unmarshal(v, &org); err != nil {
		return nil, &errors.Error{
			Err: err,
		}
	}

	return &org, nil
}
