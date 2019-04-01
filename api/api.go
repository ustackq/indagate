package api

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/prometheus/client_golang/prometheus"

	apiv1 "github.com/ustackq/indagate/api/v1"
)

// API represents APIs of indagate
type API struct {
	v1                       *apiv1.API
	requests                 prometheus.Gauge
	concurrencyLimitExceeded prometheus.Counter
	timeout                  time.Duration
	inFlightSem              chan struct{}
}

// Options for the creation of an API object. Alerts, Silences, and StatusFunc
// are mandatory to set. The zero value for everything else is a safe default.
type Options struct {
	// Timeout for all HTTP connections. The zero value (and negative
	// values) result in no timeout.
	Timeout time.Duration
	// Concurrency limit for GET requests. The zero value (and negative
	// values) result in a limit of GOMAXPROCS or 8, whichever is
	// larger. Status code 503 is served for GET requests that would exceed
	// the concurrency limit.
	Concurrency int
	// Logger is used for logging, if nil, no logging will happen.
	Logger log.Logger
	// Registry is used to register Prometheus metrics. If nil, no metrics
	// registration will happen.
	Registry prometheus.Registerer
}

// New return Api Pointer
func New(o Options) (*API, error) {
	l := o.Logger
	if l == nil {
		l = log.NewNopLogger()
	}
	concurrency := o.Concurrency
	if concurrency < 1 {
		concurrency = runtime.GOMAXPROCS(0)
		if concurrency < 8 {
			concurrency = 8
		}
	}
	v1, err := apiv1.NewAPI(l)
	v := &API{
		v1:          v1,
		timeout:     o.Timeout,
		inFlightSem: make(chan struct{}, concurrency),
	}
	return v, err
}

// Registry all APIs
func (api *API) Registry(h http.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", api.limitHandler(h))
	mux.Handle("/api/v1/",
		api.limitHandler(http.StripPrefix("/api/v1", api.v1.Handler)),
	)
	return mux
}

func (api *API) limitHandler(h http.Handler) http.Handler {
	concLimiter := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			select {
			case api.inFlightSem <- struct{}{}:
				api.requests.Inc()
				defer func() {
					<-api.inFlightSem
					api.requests.Dec()
				}()
			default:
				api.concurrencyLimitExceeded.Inc()
				http.Error(
					rw,
					fmt.Sprintf("Limit of concurrent GET requests reached (%d), try agagin later.\n", cap(api.inFlightSem)),
					http.StatusServiceUnavailable)
				return

			}
		}
		h.ServeHTTP(rw, req)
	})
	if api.timeout <= 0 {
		return concLimiter
	}
	return http.TimeoutHandler(concLimiter, api.timeout, fmt.Sprintf(
		"Exceeded configured timeout of %v.\n", api.timeout,
	))

}
