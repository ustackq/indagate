package store

import (
	"context"
	"fmt"
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

var (
	orgBucket = []byte("orgv1alpha1")
	orgIndex  = []byte("orgIndexv1alpha1")
)

func (s *Service) findOrgnizationByName(ctx context.Context, tx Impl, str string) (*service.Organization, error) {
	b, err := tx.Bucket(orgBucket)
	if err != nil {
		return nil, err
	}

	org, err := b.Get([]byte(str))
	if errors.IsNotFound(err) {
		return nil, &errors.Error{
			Code: errors.NotFound,
			Msg:  fmt.Sprintf("org %s not found", str),
		}
	}

	if err != nil {
		return nil, err
	}

	var id service.ID
	if err := id.Decode(org); err != nil {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Err:  err,
		}
	}

	return s.findOrgnizationByID(ctx, tx, id)
}

func (s *Service) findOrgnizationByID(ctx context.Context, tx Impl, id service.ID) (*service.Organization, error) {
	encodeID, err := id.Encode()
	if err != nil {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Err:  err,
			Op:   "Encode",
		}
	}

	b, err := tx.Bucket(orgBucket)
	if err != nil {
		return nil, err
	}

	v, err := b.Get(encodeID)
	if errors.IsNotFound(err) {
		return nil, &errors.Error{
			Code: errors.NotFound,
			Msg:  "org not found",
		}
	}

	if err != nil {
		return nil, err
	}

	var org *service.Organization
	if err := json.Unmarshal(v, org); err != nil {
		return nil, &errors.Error{
			Err: err,
		}
	}
	return org, nil
}
