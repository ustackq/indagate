package routes

import (
	"net/http"
	"strings"

	ihttp "github.com/ustackq/indagate/pkg/http"
)

type PlatformHandler struct {
	DocsHandler http.HandlerFunc
	APIHandler  http.Handler
}

func NewPlatformHandler(b *ihttp.APIBackend) *PlatformHandler {
	h := ihttp.NewAuthenticationHandler()
	h.Handler = ihttp.NewAPIHandler(b)
	h.AuthenticationService = b.AuthenticationService
	h.RegisterNoAuthRouter("GET", "/api/v1")
	h.RegisterNoAuthRouter("GET", "/api/v1/setup")
	h.RegisterNoAuthRouter("GET", "/api/v1/swagger.json")
	return &PlatformHandler{
		APIHandler: h,
	}
}

func (ph *PlatformHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ihttp.SetCORSResponseHeaders(rw, r)
	if r.Method == "OPTION" {
		return
	}

	// doc fisrt
	if strings.HasPrefix(r.URL.Path, "/dcos") {
		ph.DocsHandler.ServeHTTP(rw, r)
		return
	}

	ph.APIHandler.ServeHTTP(rw, r)
	return
}
