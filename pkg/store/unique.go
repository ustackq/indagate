package store

import (
	"context"
	"fmt"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

var NotUniqueError = &errors.Error{
	Code: errors.Conflict,
	Msg:  "name already exists",
}

func (s *Service) unique(ctx context.Context, tx Impl, indexBucket, indexKey []byte) error {
	bucket, err := tx.Bucket(indexBucket)
	if err != nil {
		return &errors.Error{
			Code: errors.Internal,
			Msg:  fmt.Sprintf("unexpected error gain index %v", err),
			Op:   "Index",
		}
	}

	_, err = bucket.Get(indexBucket)
	if errors.IsNotFound(err) {
		return nil
	}

	if err == nil {
		return &errors.Error{
			Code: errors.Conflict,
			Msg:  "object has existed",
		}
	}

	return &errors.Error{
		Code: errors.Internal,
		Msg:  fmt.Sprintf("unexpected error gain index %v", err),
		Op:   "Index",
	}
}
