package service

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ustackq/indagate/pkg/utils/errors"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// AuthorizationKind is returned by (*Authorization).Kind().
const AuthorizationKind = "authorization"

// auth service op
const (
	OpFindAuthorizationByID    = "FindAuthorizationByID"
	OpFindAuthorizationByToken = "FindAuthorizationByToken"
	OpFindAuthorizations       = "FindAuthorizations"
	OpCreateAuthorization      = "CreateAuthorization"
	OpUpdateAuthorization      = "UpdateAuthorization"
	OpDeleteAuthorization      = "DeleteAuthorization"
)

var (
	ErrCreateToken = &errors.Error{
		Code: errors.Invalid,
		Msg:  "unable to create token",
	}
)

// Authorization define auth object
type Authorization struct {
	ID          ID            `json:"id"`
	OrgID       ID            `json:"org"`
	Token       string        `json:"token"`
	UserID      ID            `json:"userID,omitempty"`
	Permissions []*Permission `json:"permissions"`
	Description string        `json:"description,omitempty"`
	Status      Status        `json:"active"`
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
	FindAuthorization(ctx context.Context, filter AuthorizationFilter, opt ...FindOptions) ([]*Authorization, int, error)
	CreateAuthorization(ctx context.Context, auth *Authorization) error
	UpdateAuthorization(ctx context.Context, id ID, update *AuthorizationUpdate) (*Authorization, error)
	DeleteAuthorization(ctx context.Context, id ID) error
}

// AuthorizationFilter represent a set of filter that mathch returned results.
type AuthorizationFilter struct {
	Token  *string
	ID     *ID
	UserID *ID
	User   *string
	OrgID  *ID
	Org    *string
}

func (auth *Authorization) Allowed(p Permission) bool {
	if !IsActive(auth) {
		return false
	}
	return PermissionAllowed(p, auth.Permissions)
}

func (auth *Authorization) Valid() error {
	for _, p := range auth.Permissions {
		if p.Resource.OrgID != nil && *p.Resource.OrgID != auth.OrgID {
			return &errors.Error{
				Msg:  fmt.Sprintf("permission %s is not matcher with org id %s", p, auth.OrgID),
				Code: errors.Invalid,
			}
		}
	}
	return nil
}

func IsActive(auth *Authorization) bool {
	return auth.Status == Active
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

// InstrumentedAuthorizationService
type InstrumentedAuthNService struct {
	requestCount         *prometheus.CounterVec
	requestDuration      *prometheus.HistogramVec
	AuthorizationService AuthorizationService
}

func NewAuthorizationService(a AuthorizationService) *InstrumentedAuthNService {
	namespace := "auth"
	subsystem := "prometheus"
	authn := &InstrumentedAuthNService{
		requestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "requests_total",
				Help:      "Num of http requests received",
			},
			[]string{"method", "error"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "request_duration_seconds",
				Help:      "Time taken to respond to http requests",
				Buckets:   prometheus.ExponentialBuckets(0.001, 1.5, 25),
			},
			[]string{"method", "error"},
		),
		AuthorizationService: a,
	}
	return authn
}

func (a *InstrumentedAuthNService) FindAuthorizationByID(ctx context.Context, id ID) (result *Authorization, err error) {
	defer func(start time.Time) {
		labels := prometheus.Labels{
			"method": "FindAuthorizationByID",
			"error":  fmt.Sprint(err != nil),
		}
		a.requestCount.With(labels).Add(1)
		a.requestDuration.With(labels).Observe(time.Since(start).Seconds())
	}(time.Now())
	return a.AuthorizationService.FindAuthorizationByID(ctx, id)
}

func (a *InstrumentedAuthNService) FindAuthorizationByToken(ctx context.Context, token string) (result *Authorization, err error) {
	defer func(start time.Time) {
		labels := prometheus.Labels{
			"method": "FindAuthorizationByToken",
			"error":  fmt.Sprint(err != nil),
		}
		a.requestCount.With(labels).Add(1)
		a.requestDuration.With(labels).Observe(time.Since(start).Seconds())
	}(time.Now())
	return a.AuthorizationService.FindAuthorizationByToken(ctx, token)
}

func (a *InstrumentedAuthNService) FindAuthorization(ctx context.Context, filter AuthorizationFilter, opt ...FindOptions) (result []*Authorization, n int, err error) {
	defer func(start time.Time) {
		labels := prometheus.Labels{
			"method": "FindAuthorization",
			"error":  fmt.Sprint(err != nil),
		}
		a.requestCount.With(labels).Add(1)
		a.requestDuration.With(labels).Observe(time.Since(start).Seconds())
	}(time.Now())
	return a.AuthorizationService.FindAuthorization(ctx, filter, opt...)
}

func (a *InstrumentedAuthNService) CreateAuthorization(ctx context.Context, auth *Authorization) (err error) {
	defer func(start time.Time) {
		labels := prometheus.Labels{
			"method": "CreateAuthorization",
			"error":  fmt.Sprint(err != nil),
		}
		a.requestCount.With(labels).Add(1)
		a.requestDuration.With(labels).Observe(time.Since(start).Seconds())
	}(time.Now())
	return a.AuthorizationService.CreateAuthorization(ctx, auth)
}

func (a *InstrumentedAuthNService) UpdateAuthorization(ctx context.Context, id ID, update *AuthorizationUpdate) (err error) {
	defer func(start time.Time) {
		labels := prometheus.Labels{
			"method": "UpdateAuthorization",
			"error":  fmt.Sprint(err != nil),
		}
		a.requestCount.With(labels).Add(1)
		a.requestDuration.With(labels).Observe(time.Since(start).Seconds())
	}(time.Now())
	return a.UpdateAuthorization(ctx, id, update)
}

func (a *InstrumentedAuthNService) DeleteAuthorization(ctx context.Context, id ID) (err error) {
	defer func(start time.Time) {
		labels := prometheus.Labels{
			"method": "DeleteAuthorization",
			"error":  fmt.Sprint(err != nil),
		}
		a.requestCount.With(labels).Add(1)
		a.requestDuration.With(labels).Observe(time.Since(start).Seconds())
	}(time.Now())
	return a.AuthorizationService.DeleteAuthorization(ctx, id)
}
