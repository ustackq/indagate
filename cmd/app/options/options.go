package options

import (
	"io"
	"sync"
	"time"

	
	"go.uber.org/zap"

	"github.com/ustackq/indagate/pkg/http"
	"github.com/ustackq/indagate/pkg/metrics"
	"github.com/ustackq/indagate/pkg/nats"
	"github.com/ustackq/indagate/config"
)

// StoreConfig define store config
type StoreConfig struct {
	Type    string
	Host    string
	Name    string
	User    string
	Paaswd  string
	Path    string
	SSLMode string
}

// Indagate contains configuration flags for the Indagate.
type Indagate struct {
	Config  *config.Config
	wg      sync.WaitGroup
	running bool
	cancel  func()
	// define testing wether
	testing bool
	// storeConfig means the kind of store,now supported:mysql、postsql
	storeConfig StoreConfig
	// secretConfig define the kind of store, now supported:mysql、vault
	secretConfig StoreConfig
	// tracingType define app tracing type: now supported: opentracing、opencensus
	tracingType string
	// server address
	addr string
	// Queue config
	// TODO: supported other mq
	natsServer *nats.Server
	// define graceful stop timeout
	Timeout time.Duration
	Logger  *zap.Logger
	metric  *metrics.Registry
	// ouput
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

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

func NewIndagateOptions(config *config.Config) *Indagate {
	return &Indagate{
		Config: config,
	}
}