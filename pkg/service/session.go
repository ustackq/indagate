package service

import (
	"context"
	"github.com/ustackq/indagate/pkg/utils/errors"
	"time"
)

// ErrSessionExpired is the error message for expired sessions.
const ErrSessionExpired = "session has expired"

// RenewSessionTime is the the time to extend session, currently set to 5min.
var RenewSessionTime = time.Duration(time.Second * 300)

// SessionAuthorizionKind defines the type of authorizer
const SessionAuthorizionKind = "session"

// Session represents user session
type Session struct {
	ID          ID            `json:"id"`
	Key         string        `json:"key"`
	CreatedAt   time.Time     `json:"createAt"`
	ExpiresAt   time.Time     `json:"expiresAt"`
	UserID      ID            `json:"userID,omitempty"`
	Permissions []*Permission `json:"permissions,omitempty"`
}

type SessionService interface {
	FindSession(ctx context.Context, key string) (*Session, error)
	ExpireSession(ctx context.Context, key string) error
	CreateSession(ctx context.Context, user string) (*Session, error)
	RenewSession(ctx context.Context, session *Session, newExpiration time.Time) error
}

func (s *Session) Expired() error {
	if time.Now().After(s.ExpiresAt) {
		return &errors.Error{
			Code: errors.Forbidden,
			Msg:  ErrSessionExpired,
		}
	}
	return nil
}

// Allowed return true if the author is unexpired and request permission exists in the sessions list
func (s *Session) Allowed(p Permission) bool {
	if err := s.Expired(); err != nil {
		return false
	}
	return PermissionAllowed(p, s.Permissions)
}

// Kind return the kind of auditing.
func (s *Session) Kind() string {
	return SessionAuthorizionKind
}

// Identifier returns the session id which used to auditing.
func (s *Session) Identifier() ID {
	return s.ID
}

// GetUserID return the user id.
func (s *Session) GetUserID() ID {
	return s.UserID
}
