package authorizer

import (
	"context"
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/tracing"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

var _ service.BucketService = (*BucketService)(nil)

// BucketService wraps a influxdb.BucketService and authorizes actions
// against it appropriately.
type BucketService struct {
	s service.BucketService
}

// NewBucketService constructs an instance of an authorizing bucket serivce.
func NewBucketService(s service.BucketService) *BucketService {
	return &BucketService{
		s: s,
	}
}

func newBucketPermission(a service.Action, orgID, id service.ID) (*service.Permission, error) {
	return service.NewPermissionAtID(id, a, service.BucketsResourceType, orgID)
}

func authorizeReadBucket(ctx context.Context, orgID, id service.ID) error {
	span, ctx := tracing.StartSpanFromContext(ctx)
	defer span.End()

	p, err := newBucketPermission(service.ReadAction, orgID, id)
	if err != nil {
		return err
	}

	if err := isAllowed(ctx, *p); err != nil {
		return err
	}

	return nil
}

func authorizeWriteBucket(ctx context.Context, orgID, id service.ID) error {
	p, err := newBucketPermission(service.WriteAction, orgID, id)
	if err != nil {
		return err
	}

	if err := isAllowed(ctx, *p); err != nil {
		return err
	}

	return nil
}

// FindBucketByID checks to see if the authorizer on context has read access to the id provided.
func (s *BucketService) FindBucketByID(ctx context.Context, id service.ID) (*service.Bucket, error) {
	span, ctx := tracing.StartSpanFromContext(ctx)
	defer span.End()

	b, err := s.s.FindBucketByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := authorizeReadBucket(ctx, b.OrgID, id); err != nil {
		return nil, err
	}

	return b, nil
}

// FindBucket retrieves the bucket and checks to see if the authorizer on context has read access to the bucket.
func (s *BucketService) FindBucket(ctx context.Context, filter service.BucketFilter) (*service.Bucket, error) {
	span, ctx := tracing.StartSpanFromContext(ctx)
	defer span.End()

	b, err := s.s.FindBucket(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err := authorizeReadBucket(ctx, b.OrgID, b.ID); err != nil {
		return nil, err
	}

	return b, nil
}

// FindBuckets retrieves all buckets that match the provided filter and then filters the list down to only the resources that are authorized.
func (s *BucketService) FindBuckets(ctx context.Context, filter service.BucketFilter, opt ...service.FindOptions) ([]*service.Bucket, int, error) {
	span, ctx := tracing.StartSpanFromContext(ctx)
	defer span.End()

	// TODO: we'll likely want to push this operation into the database eventually since fetching the whole list of data
	// will likely be expensive.
	bs, _, err := s.s.FindBuckets(ctx, filter, opt...)
	if err != nil {
		return nil, 0, err
	}

	// This filters without allocating
	// https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
	buckets := bs[:0]
	for _, b := range bs {
		err := authorizeReadBucket(ctx, b.OrgID, b.ID)
		if err != nil && errors.ErrorCode(err) != errors.Unauthorized {
			return nil, 0, err
		}

		if errors.ErrorCode(err) == errors.Unauthorized {
			continue
		}

		buckets = append(buckets, b)
	}

	return buckets, len(buckets), nil
}

// CreateBucket checks to see if the authorizer on context has write access to the global buckets resource.
func (s *BucketService) CreateBucket(ctx context.Context, b *service.Bucket) error {
	span, ctx := tracing.StartSpanFromContext(ctx)
	defer span.End()

	p, err := service.NewPermission(service.WriteAction, service.BucketsResourceType, b.OrgID)
	if err != nil {
		return err
	}

	if err := isAllowed(ctx, *p); err != nil {
		return err
	}

	return s.s.CreateBucket(ctx, b)
}

// UpdateBucket checks to see if the authorizer on context has write access to the bucket provided.
func (s *BucketService) UpdateBucket(ctx context.Context, id service.ID, upd service.BucketUpdate) (*service.Bucket, error) {
	b, err := s.s.FindBucketByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := authorizeWriteBucket(ctx, b.OrgID, id); err != nil {
		return nil, err
	}

	return s.s.UpdateBucket(ctx, id, upd)
}

// DeleteBucket checks to see if the authorizer on context has write access to the bucket provided.
func (s *BucketService) DeleteBucket(ctx context.Context, id service.ID) error {
	b, err := s.s.FindBucketByID(ctx, id)
	if err != nil {
		return err
	}

	if err := authorizeWriteBucket(ctx, b.OrgID, id); err != nil {
		return err
	}

	return s.s.DeleteBucket(ctx, id)
}
