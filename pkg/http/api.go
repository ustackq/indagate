package http

import (
	"net/http"
	"strings"

	"github.com/ustackq/indagate/pkg/authorizer"
	"github.com/ustackq/indagate/pkg/service"
	"go.uber.org/zap"
)

// APIHandler is a collection of all service handlers.
type APIHandler struct {
	SessionHandler       *SessionHandler
	OrgHandler           *OrgHandler
	UserHandler          *UserHandler
	SetupHandler         *SetupHandler
	AuthorizationHandler *AuthorizationHandler
	SwaggerHandler       http.Handler
}

// APIBackend is used to constrcted service handlers.
type APIBackend struct {
	// if empty then assets ar served from bindata.
	AssetPath            string
	Logger               *zap.Logger
	SessionRenewDisabled bool

	PasswordsService           service.PasswordsService
	BucketService              service.BucketService
	SetupService               service.SetupService
	AuthenticationService      service.AuthorizationService
	SessionService             service.SessionService
	UserService                service.UserService
	UserResourceMappingService service.UserResourceMappingService
	OrganizationService        service.OrganizationService
	LookupService              service.LookupService
	OrgLookupService           authorizer.OrganizationService
}

// NewAPIHandler construct APIHandler
func NewAPIHandler(ab *APIBackend) *APIHandler {
	ah := &APIHandler{}
	// tmp UserMappingResource
	ab.UserResourceMappingService = authorizer.NewUserMappingService(ab.OrgLookupService, ab.UserResourceMappingService)

	// create session handler
	sessionBackend := NewSessionBackend(ab)
	ah.SessionHandler = NewSessionHandler(sessionBackend)

	// create bucket handler

	// create org handler
	orgBackend := NewOrgBackend(ab)
	orgBackend.OrganizationService = authorizer.NewOrgService(ab.OrganizationService)
	ah.OrgHandler = NewOrgHandler(orgBackend)

	// create user handler
	userBackend := NewUserBackend(ab)
	userBackend.UserService = authorizer.NewUserService(ab.UserService)
	ah.UserHandler = NewUserHandler(userBackend)

	// create authorization handler
	authorizationBackend := NewAuthorizationBackend(ab)
	authorizationBackend.AuthorizationService = authorizer.NewAuthorizationService(ab.AuthenticationService)
	ah.AuthorizationHandler = NewAuthorizationHandler(authorizationBackend)

	stb := NewSetupBackend(ab)
	ah.SetupHandler = NewSetupHandler(stb)
	ah.SwaggerHandler = newSwaggerLoader(stb.Logger.With(zap.String("SERVICE", "swagger-loader")))
	return ah
}

var api = map[string]interface{}{
	"authorizations": "/api/v1/authorizations",
	"buckets":        "/api/v1/buckets",
	"me":             "/api/v1/me",
	"orgs":           "/api/v1/orgs",
	"setup":          "/api/v1/setup",
	"signin":         "/api/v1/signin",
	"signout":        "/api/v1/signout",
	"system": map[string]string{
		"metrics": "/metrics",
		"debug":   "/debug/pprof",
		"health":  "/healthz",
	},
	"swagger": "/api/v1/swagger.json",
}

func (ah *APIHandler) serveLinks(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := encodeResponse(ctx, rw, http.StatusOK, api); err != nil {
		EncodeError(ctx, err, rw)
		return
	}
}

func (ah *APIHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	SetCORSResponseHeaders(rw, r)
	if r.Method == "OPTIONS" {
		return
	}

	if r.URL.Path == "/api/v1/" || r.URL.Path == "/api/v1" {
		ah.serveLinks(rw, r)
		return
	}

	if r.URL.Path == "/api/v1/signin" || r.URL.Path == "/api/v1/signout" {
		ah.SessionHandler.ServeHTTP(rw, r)
		return
	}

	if r.URL.Path == "/api/v2/swagger.json" {
		ah.SwaggerHandler.ServeHTTP(rw, r)
		return
	}

	if r.URL.Path == "/api/v1/setup" {
		ah.SetupHandler.ServeHTTP(rw, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v1/users") {
		ah.UserHandler.ServeHTTP(rw, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v1/me") {
		ah.UserHandler.ServeHTTP(rw, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v1/authorizations") {
		ah.AuthorizationHandler.ServeHTTP(rw, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/v1/orgs") {
		ah.OrgHandler.ServeHTTP(rw, r)
		return
	}

	notFoundHandler(rw, r)
}
