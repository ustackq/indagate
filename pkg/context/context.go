package context

import (
	"context"
	"fmt"

	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

type contextKey string

const (
	authzCtxKey = contextKey("indagate/authorizer/v1")
)

// SetAuthorizer sets an authorizer on context.
func SetAuthorizer(ctx context.Context, authz service.Authorizer) context.Context {
	return context.WithValue(ctx, authzCtxKey, authz)
}

func GetAuthorizer(ctx context.Context) (service.Authorizer, error) {
	authz, ok := ctx.Value(authzCtxKey).(service.Authorizer)
	if !ok {
		return nil, &errors.Error{
			Msg:  "Authotizer not found on context",
			Code: errors.Internal,
		}
	}
	return authz, nil
}

func GetToken(ctx context.Context) (string, error) {
	authz, ok := ctx.Value(authzCtxKey).(service.Authorizer)
	if !ok {
		return "", &errors.Error{
			Msg:  "Authotizer not found on context",
			Code: errors.Internal,
		}
	}
	auth, ok := authz.(*service.Authorization)
	if !ok {
		return "", &errors.Error{
			Msg:  fmt.Sprintf("Authorizer not an authorization but %T", auth),
			Code: errors.Internal,
		}
	}
	return auth.Token, nil
}
