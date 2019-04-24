package options

import (
	"context"
	"io"
	nethttp "net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ustackq/indagate/config"
	"github.com/ustackq/indagate/pkg/http"
	"github.com/ustackq/indagate/pkg/metrics"
	"github.com/ustackq/indagate/pkg/nats"
)

// Indagate contains configuration flags for the Indagate.
type Indagate struct {
	// Config define configuration file
	Config  string
	wg      sync.WaitGroup
	running bool
	// used for testing
	cancel func()
	// define testing wether or not
	testing bool
	// storeConfig means the kind of store,now supported:mysql、postsql
	storeConfig config.Store
	// secretConfig define the kind of store, now supported:mysql、vault
	secretConfig config.Store
	// tracingType define app tracing type: now supported: opentracing、opencensus
	tracingType string
	telemetry   bool
	// Queue config
	// TODO: supported other mq
	natsServer *nats.Server
	// define graceful stop timeout
	Timeout  time.Duration
	Logger   *zap.Logger
	logLevel string
	metric   *metrics.Registry
	// ouput
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	httpAddress struct {
		addr string
		port string
	}
	server nethttp.Server
	// Backend service
	backend *http.APIBackend
}

// NewIndagateFlags will create a new IndagateFlags with default values.
func NewIndagateFlags() *Indagate {
	return &Indagate{}
}

// ValidataIndagate validate Indagate flag config
func ValidataIndagate(ing *Indagate) error {
	return nil
}

func NewIndagateOptions(config string) *Indagate {
	return &Indagate{
		Config: config,

		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

func (ing *Indagate) Shutdown(ctx context.Context) {
	ing.server.Shutdown(ctx)

	ing.Logger.Info("Shutting donw", zap.String("service", "nats"))
	ing.natsServer.Close()
	ing.Logger.Sync()
}
