package metrics

import (
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"net/http"
)

// Registry wrap prometheus registry
type Registry struct {
	*prom.Registry
	logger *zap.Logger
}

func NewRegistry() *Registry {
	return &Registry{
		Registry: prom.NewRegistry(),
		logger:   zap.NewNop(),
	}
}

func (reg *Registry) WithLogger(logger *zap.Logger) {
	reg.logger = logger.With(zap.String("service", "registry"))
}

func (reg *Registry) Handler() http.Handler {
	opts := promhttp.HandlerOpts{
		ErrorLog: promLogger{r: reg},
		// TODO(mr): decide if we want to set MaxRequestsInFlight or Timeout.
	}
	return promhttp.HandlerFor(reg.Registry, opts)
}

// promLogger satisfies the promhttp.Logger interface with the registry.
// Because normal usage is that WithLogger is called after HTTPHandler,
// we refer to the Registry rather than its logger.
type promLogger struct {
	r *Registry
}

var _ promhttp.Logger = (*promLogger)(nil)

// Println implements promhttp.Logger.
func (pl promLogger) Println(v ...interface{}) {
	pl.r.logger.Sugar().Info(v...)
}
