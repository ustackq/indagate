package http

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"net/http"
	"strings"

	"github.com/ustackq/indagate/pkg/metrics"
)

const (
	// MetricsPath exposes prometheus format metrics via /metrics
	MetricsPath = "/metrics"
	// StatusPath exposes the status of the service via /status which represent service info
	StatusPath = "/status"
	// HealthPath exposes the health of the service via /health
	HealthPath = "/health"
	// DebugPath exposes debugging info via /debug/pprof
	DebugPath = "/debug"
)

// Handler provides standard handling of metrics, health, status and debug endpoints.
// All other requests are serveing via sub handler.
type Handler struct {
	name string
	// MetricsHandler handles metrics endponit
	MetricsHandler http.Handler
	// StatusPath handles sttaus info endpoint
	StatusHandler http.Handler
	// HealthHandler handles service's health endpoint
	HealthPath http.Handler
	// DebugHandler handles debug endpoint
	DebugPath http.Handler
	// Handler handles all others requests
	Handler http.Handler

	requests   *prometheus.CounterVec
	requestDur *prometheus.HistogramVec

	// Logger if set will log all requests as served
	Logger *zap.Logger
}

func encodeResponse(ctx context.Context, rw http.ResponseWriter, code int, res interface{}) error {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(code)
	return json.NewEncoder(rw).Encode(res)
}

func NewHandlerWithRegistry(name string, reg *metrics.Registry) *Handler {
	h := &Handler{
		name:           name,
		MetricsHandler: reg.Handler(),
		StatusHandler:  http.HandlerFunc(StatusHandler),
		HealthPath:     http.HandlerFunc(HealthHandler),
		DebugPath:      http.DefaultServeMux,
	}
	h.initMetrics()
	reg.MustRegister(h.PrometheusCollectors()...)
	return h
}

func (h *Handler) PrometheusCollectors() []prometheus.Collector {
	return []prometheus.Collector{
		h.requests,
		h.requestDur,
	}
}

func (h *Handler) initMetrics() {
	const namespace = "http"
	const handlerSubsystem = "api"

	h.requests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: handlerSubsystem,
		Name:      "requests_total",
		Help:      "Number of http requests received",
	}, []string{"handler", "method", "path", "status", "user_agent"})

	h.requestDur = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: handlerSubsystem,
		Name:      "request_duration_seconds",
		Help:      "Time taken to respond to HTTP request",
	}, []string{"handler", "method", "path", "status", "user_agent"})
}

// ServerHTTP implement Handler interface
func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// TODO tracing
	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		userAgent = "UNKNOWN"
	}

	// TODO: how apiserver and influxdb to handle metrics

	switch {
	case r.URL.Path == MetricsPath:
		h.MetricsHandler.ServeHTTP(rw, r)
	case r.URL.Path == StatusPath:
		h.StatusHandler.ServeHTTP(rw, r)
	case strings.HasPrefix(r.URL.Path, DebugPath):
		h.DebugPath.ServeHTTP(rw, r)
	default:
		h.Handler.ServeHTTP(rw, r)
	}

}
