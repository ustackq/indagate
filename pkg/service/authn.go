package service

import "context"

// AuthorizationKind is returned by (*Authorization).Kind().
const AuthorizationKind = "authorization"

// Authorization define auth object
type Authorization struct {
	ID         ID           `json:"id"`
	Token      string       `json:"token"`
	Active     bool         `json:"active"`
	UserID     ID           `json:"userID,omitempty"`
	Permission []Permission `json:"permissions"`
}

// AuthorizationUpdate define update object
type AuthorizationUpdate struct {
	Status      *Status `json:"status"`
	Description *string `json:"description,omitempty"`
}

// AuthorizationService represents a service which provider authorization service.
type AuthorizationService interface {
	FindAuthorizationByID(ctx context.Context, id ID) (*Authorization, error)
	FindAuthorizationByToken(ctx context.Context, token string) (*Authorization, error)
	FindAuthorization(ctx context.Context, filter AuthorizationFilter, opt ...FindOptions) ([]*Authorization, error)
	CreateAuthorization(ctx context.Context) (*Authorization, error)
	UpdateAuthorization(ctx context.Context, id ID, update *AuthorizationUpdate) error
	DeleteAuthorization(ctx context.Context, id ID) error
}

// AuthorizationFilter represent a set of filter that mathch returned results.
type AuthorizationFilter struct {
	Token *string
	ID    *ID
}

func (auth *Authorization) Allowed(p Permission) bool {
	if !IsActive(auth) {
		return false
	}
	return PermissionAllowed(p, auth.Permission)
}

func IsActive(auth *Authorization) bool {
	return auth.Active == true
}

func (auth *Authorization) Kind() string {
	return AuthorizationKind
}

func (auth *Authorization) Identifier() ID {
	return auth.ID
}

func (auth *Authorization) GetUserID() ID {
	return auth.UserID
}
