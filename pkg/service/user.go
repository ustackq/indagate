package service

import (
	"context"
)

// User define a user info
type User struct {
	ID   ID     `json:"id,omitempty"`
	Name string `json:"name"`
}

type UserFilter struct {
	ID   *ID
	Name *string
}

type UserUpdate struct {
	Name *string `json:"name"`
}

type UserService interface {
	FindUserByID(ctx context.Context, id ID) (*User, error)
	FindUser(ctx context.Context, filter UserFilter) (*User, error)
	FindUsers(ctx context.Context, filter UserFilter, opts ...FindOptions) ([]*User, int, error)
	CreateUser(ctx context.Context, u *User) error

	UpdateUser(ctx context.Context, id ID, update UserUpdate) (*User, error)
	DeleteUser(ctx context.Context, id ID) error
}
