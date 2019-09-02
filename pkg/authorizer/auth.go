package authorizer

import (
	"context"
	"fmt"
	icontext "github.com/ustackq/indagate/pkg/context"
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

var _ service.AuthorizationService = (*AuthorizationService)(nil)

// AuthorizationService wraps the AuthorizationService and authorizes actions
// agaist it appropriately.
type AuthorizationService struct {
	s service.AuthorizationService
}

func NewAuthorizationService(s service.AuthorizationService) *AuthorizationService {
	return &AuthorizationService{
		s: s,
	}
}

func newAuthorizationPermission(action service.Action, id service.ID) (*service.Permission, error) {
	p := &service.Permission{
		Action: action,
		Resource: service.Resource{
			Type: service.UsersResourceType,
			ID:   &id,
		},
	}

	return p, p.Valid()
}

func authorizaAuthZByAction(action service.Action, ctx context.Context, id service.ID) error {
	p, err := newAuthorizationPermission(action, id)
	if err != nil {
		return err
	}

	if err := isAllowed(ctx, *p); err != nil {
		return err
	}

	return nil
}

func (s *AuthorizationService) FindAuthorizationByID(ctx context.Context, id service.ID) (*service.Authorization, error) {
	a, err := s.s.FindAuthorizationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := authorizaAuthZByAction(service.ReadAction, ctx, a.UserID); err != nil {
		return nil, err
	}

	return a, nil
}

func (s *AuthorizationService) FindAuthorizationByToken(ctx context.Context, token string) (*service.Authorization, error) {
	a, err := s.s.FindAuthorizationByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if err := authorizaAuthZByAction(service.ReadAction, ctx, a.UserID); err != nil {
		return nil, err
	}

	return a, nil
}

func (s *AuthorizationService) FindAuthorization(ctx context.Context, filter service.AuthorizationFilter, opts ...service.FindOptions) ([]*service.Authorization, int, error) {
	// TODO: put results into databases which expensive
	auths, _, err := s.s.FindAuthorization(ctx, filter, opts...)
	if err != nil {
		return nil, 0, err
	}

	// need understand https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
	as := auths[:0]
	for _, a := range auths {
		err := authorizaAuthZByAction(service.ReadAction, ctx, a.UserID)
		if err != nil && errors.ErrorCode(err) != errors.Unauthorized {
			return nil, 0, err
		}

		if errors.ErrorCode(err) == errors.Unauthorized {
			continue
		}

		as = append(as, a)
	}

	return as, len(as), nil
}

func (s *AuthorizationService) CreateAuthorization(ctx context.Context, a *service.Authorization) error {
	if err := authorizaAuthZByAction(service.ReadAction, ctx, a.UserID); err != nil {
		return err
	}

	if err := verifyPermissions(ctx, a.Permissions); err != nil {
		return err
	}

	return s.s.CreateAuthorization(ctx, a)

}

func verifyPermissions(ctx context.Context, ps []*service.Permission) error {
	a, err := icontext.GetAuthorizer(ctx)
	if err != nil {
		return err
	}

	for _, p := range ps {
		if !a.Allowed(*p) {
			return &errors.Error{
				Err: &errors.Error{
					Code: errors.Unauthorized,
					Msg:  fmt.Sprintf("%s is unauthorized", p),
				},
				Msg:  fmt.Sprintf("permission %s is not allowded", p),
				Code: errors.Forbidden,
			}
		}
	}
	return nil
}

func (s *AuthorizationService) UpdateAuthorization(ctx context.Context, id service.ID, update *service.AuthorizationUpdate) (*service.Authorization, error) {
	a, err := s.s.FindAuthorizationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := authorizaAuthZByAction(service.WriteAction, ctx, a.UserID); err != nil {
		return nil, err
	}

	return s.s.UpdateAuthorization(ctx, id, update)
}

// DeleteAuthorization checks to see if the authorizer on context has write access to the authorization provided.
func (s *AuthorizationService) DeleteAuthorization(ctx context.Context, id service.ID) error {
	a, err := s.s.FindAuthorizationByID(ctx, id)
	if err != nil {
		return err
	}

	if err := authorizaAuthZByAction(service.WriteAction, ctx, a.UserID); err != nil {
		return err
	}

	return s.s.DeleteAuthorization(ctx, id)
}
