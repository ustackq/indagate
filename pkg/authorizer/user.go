package authorizer

import (
	"context"
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

var _ service.UserService = (*UserService)(nil)

// UserService wraps service.UserService for authorizes actions
type UserService struct {
	s service.UserService
}

func NewUserService(s service.UserService) *UserService {
	return &UserService{
		s: s,
	}
}

func newUserPermission(action service.Action, id service.ID) (*service.Permission, error) {
	p := &service.Permission{
		Action: action,
		Resource: service.Resource{
			Type: service.UsersResourceType,
			ID:   &id,
		},
	}

	return p, p.Valid()
}

func authorizeUserByAction(action service.Action, ctx context.Context, id service.ID) error {
	p, err := newUserPermission(action, id)
	if err != nil {
		return err
	}

	if err := isAllowed(ctx, *p); err != nil {
		return err
	}

	return nil
}

func (u *UserService) FindUserByID(ctx context.Context, id service.ID) (*service.User, error) {
	if err := authorizeUserByAction(service.ReadAction, ctx, id); err != nil {
		return nil, err
	}

	return u.s.FindUserByID(ctx, id)
}

func (u *UserService) FindUser(ctx context.Context, filter service.UserFilter) (*service.User, error) {
	user, err := u.s.FindUser(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err := authorizeUserByAction(service.ReadAction, ctx, user.ID); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserService) FindUsers(ctx context.Context, filter service.UserFilter, opts ...service.FindOptions) ([]*service.User, int, error) {
	// TODO: cache result
	us, _, err := u.s.FindUsers(ctx, filter, opts...)
	if err != nil {
		return nil, 0, err
	}

	users := us[:0]
	for _, user := range us {
		err := authorizeUserByAction(service.ReadAction, ctx, user.ID)
		if err != nil && errors.ErrorCode(err) != errors.Unauthorized {
			return nil, 0, err
		}

		if errors.ErrorCode(err) == errors.Unauthorized {
			continue
		}

		users = append(users, user)
	}

	return users, len(users), nil
}

// CreateUser checks to see if the authorizer on context has write access to the global users resource.
func (u *UserService) CreateUser(ctx context.Context, o *service.User) error {
	p, err := service.NewGlobalPermission(service.WriteAction, service.UsersResourceType)
	if err != nil {
		return err
	}

	if err := isAllowed(ctx, *p); err != nil {
		return err
	}

	return u.s.CreateUser(ctx, o)
}

// UpdateUser checks to see if the authorizer on context has write access to the user provided.
func (u *UserService) UpdateUser(ctx context.Context, id service.ID, update service.UserUpdate) (*service.User, error) {
	if err := authorizeUserByAction(service.WriteAction, ctx, id); err != nil {
		return nil, err
	}

	return u.s.UpdateUser(ctx, id, update)
}

// DeleteUser checks to see if the authorizer on context has write access to the user provided.
func (u *UserService) DeleteUser(ctx context.Context, id service.ID) error {
	if err := authorizeUserByAction(service.WriteAction, ctx, id); err != nil {
		return err
	}

	return u.s.DeleteUser(ctx, id)
}
