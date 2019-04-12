package http

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ustackq/indagate/pkg/service"
	"go.uber.org/zap"
)

const (
	tokenAuthScheme = "token"
)

func SetCORSResponseHeaders(rw http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
	}
}

// AuthenticationHandler is a middlerware of authenticating incoming requests.
type AuthenticationHandler struct {
	Logger                *zap.Logger
	AuthenticationService service.AuthorizationService
	// use to lookup handler
	noAuthRouter *httprouter.Router
	Handler      http.Handler
}

// NewAuthenticationHandler retrun a new instance
func NewAuthenticationHandler() *AuthenticationHandler {
	return &AuthenticationHandler{
		Logger:       zap.NewNop(),
		Handler:      http.DefaultServeMux,
		noAuthRouter: httprouter.New(),
	}
}

// RegisterNoAuthRouter handle routes with authentication
func (ah *AuthenticationHandler) RegisterNoAuthRouter(method, path string) {
	ah.noAuthRouter.HandlerFunc(method, path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}

func (ah *AuthenticationHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if handler, _, _ := ah.noAuthRouter.Lookup(r.Method, r.URL.Path); handler != nil {
		ah.Handler.ServeHTTP(rw, r)
		return
	}
	ctx := r.Context()
	scheme, err := ProbeAuthScheme(r)
	if err != nil {
		UnauthorizedError(ctx, rw)
		return
	}
	switch scheme {
	case tokenScheme:
		ctx, err = ah.extractAuthorization(ctx, r)
		if err != nil {
			break
		}
		r = r.WithContext(ctx)
		ah.Handler.ServeHTTP(rw, r)
		return
	}
}

func (ah *AuthenticationHandler) extractAuthorization(ctx context.Context, r *http.Request) (context.Context, error) {
	token, err := GetToken(r)
	if err != nil {
		return nil, err
	}

	_, err = ah.AuthenticationService.FindAuthorizationByToken(ctx, token)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}
