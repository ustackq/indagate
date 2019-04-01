package v1

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/rs/cors"
	"github.com/ustackq/indagate/api/v1/restapi"
	"github.com/ustackq/indagate/api/v1/restapi/operations"
	"github.com/ustackq/indagate/config"
)

// API provides registration of handlers for API routes.
type API struct {
	mtx    sync.RWMutex
	config *config.Config
	uptime time.Time
	logger log.Logger

	Handler http.Handler
}

// NewAPI return API Pointer
func NewAPI(l log.Logger) (*API, error) {
	api := API{
		logger: l,
		uptime: time.Now(),
	}
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded swagger file: %v", err.Error())
	}
	// create new service API
	openAPI := operations.NewIndagateAPI(swaggerSpec)

	openAPI.Middleware = func(b middleware.Builder) http.Handler {
		return middleware.Spec("", swaggerSpec.Raw(), openAPI.Context().RoutesHandler(b))
	}

	openAPI.Logger = func(s string, i ...interface{}) {
		level.Error(api.logger).Log(i...)
	}
	handleCORS := cors.Default().Handler
	api.Handler = handleCORS(openAPI.Serve(nil))

	return &api, nil
}
