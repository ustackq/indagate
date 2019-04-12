package http

import (
	"net/http"

	"github.com/ustackq/indagate/pkg/service"
	"go.uber.org/zap"
)

// APIHandler is a collection of all service handlers.
type APIHandler struct {
	SetupHandler   *SetupHandler
	SwaggerHandler http.Handler
}

// APIBackend is used to constrcted service handlers.
type APIBackend struct {
	Logger                *zap.Logger
	SetupService          service.SetupService
	AuthenticationService service.AuthorizationService
}

// NewAPIHandler construct APIHandler
func NewAPIHandler(ab *APIBackend) *APIHandler {
	ah := &APIHandler{}
	stb := NewSetupBackend(ab)
	ah.SetupHandler = NewSetupHandler(stb)
	ah.SwaggerHandler = newSwaggerLoader(stb.Logger.With(zap.String("SERVICE", "swagger-loader")))
	return ah
}

var api = map[string]interface{}{
	"setup":   "/api/v1/setup",
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

	if r.URL.Path == "/api/v2/swagger.json" {
		ah.SwaggerHandler.ServeHTTP(rw, r)
		return
	}

	if r.URL.Path == "/api/v1/setup" {
		ah.SetupHandler.ServeHTTP(rw, r)
		return
	}

}
