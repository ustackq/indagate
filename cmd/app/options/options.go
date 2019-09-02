package options

import (
	"context"
	"fmt"
	"io"
	"net"
	nethttp "net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ustackq/indagate/config"
	"github.com/ustackq/indagate/pkg/http"
	"github.com/ustackq/indagate/pkg/logger"
	"github.com/ustackq/indagate/pkg/metrics"
	"github.com/ustackq/indagate/pkg/nats"
	"github.com/ustackq/indagate/pkg/server"
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/store"
	"github.com/ustackq/indagate/pkg/store/bolt"
	"github.com/ustackq/indagate/pkg/tracing"
	"github.com/ustackq/indagate/pkg/version"
	"github.com/ustackq/indagate/routes"
)

const (
	// JaegerTracing enables tracing via the Jaeger client library
	JaegerTracing = "jaeger"
	// opencensus enables traing via opencensus client library
	OpencensusTracing = "opencensus"
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
	// define store type: means the kind of store,now supported:mysql、postsql
	storeType string
	// secretType define the type of secret store
	secretType string
	// bolt config
	boltClient *bolt.Client
	boltPath   string
	// storeConfig means the kind of store,now supported:mysql、postsql
	storeService *store.Service
	// sessionLength define session store time
	sessionLength int64
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
		port int
	}
	register *metrics.Registry
	server   nethttp.Server
	// Backend service
	jaegerTracerCloser io.Closer
	backend            *http.APIBackend
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

	ing.Logger.Info("Shutting down", zap.String("service", "nats"))
	ing.natsServer.Close()
	ing.wg.Wait()

	if ing.jaegerTracerCloser != nil {
		if err := ing.jaegerTracerCloser.Close(); err != nil {
			ing.Logger.Warn("failed to close jaeger tracer", zap.Error(err))
		}
	}
	ing.Logger.Sync()
}

// Parse parse cfg to build indagate
func (ing *Indagate) Parse(cfg string) {
	if cfg == "" {
		return
	}
	var conf config.Configuration
	if f, err := os.Lstat(cfg); !f.Mode().IsRegular() || err != nil {
		fmt.Fprintln(ing.Stderr, err)
		os.Exit(1)
	}
	f, err := os.Open(cfg)
	if err != nil {
		fmt.Fprintln(ing.Stderr, err)
		os.Exit(1)
	}
	if err = yaml.NewDecoder(f).Decode(&conf); err != nil {
		fmt.Fprintln(ing.Stderr, err)
		os.Exit(1)
	}
	lconf := &logger.Config{
		Format: "auto",
		Level:  conf.Loglevel,
	}
	log, err := lconf.New(ing.Stdout)
	if err != nil {
		fmt.Fprintln(ing.Stderr, err)
		os.Exit(1)
	}

	ing.Logger = log
}

func (ing *Indagate) SecretStore() config.Store {
	return ing.secretConfig
}

func (ing *Indagate) TelemetryEnabled() bool {
	return ing.telemetry
}

func (ing *Indagate) Registry() *metrics.Registry {
	return ing.register
}

func (ing *Indagate) Run(ctx context.Context) (err error) {
	// start tracing
	// TODO: complete tracing
	// consider apiserver tracing
	span, ctx := tracing.StartSpanFromContext(ctx)
	defer span.End()
	// set indagate server state: running
	ing.running = true
	// constrcut context
	ctx, ing.cancel = context.WithCancel(ctx)

	var level zapcore.Level
	if err := level.Set(ing.logLevel); err != nil {
		return fmt.Errorf("invalid log level; only supported DEBUG, INFO, and ERROR")
	}

	// build logger conf
	// understand the reason of logger
	logConf := &logger.Config{
		Format: "auto",
		Level:  level,
	}
	ing.Logger, err = logConf.New(os.Stdout)
	if err != nil {
		return err
	}

	// build version
	info := version.Get()
	ing.Logger.Info("Welcome to Indagate",
		zap.String("Version", info.GitVersion),
		zap.String("commit", info.GitCommit),
		zap.String("BuildDate", info.BuildDate),
	)

	// config tracing
	switch ing.tracingType {
	case OpencensusTracing:
		ing.Logger.Info("tracing via Census")
		// sth need to be done here.
	}

	// TODO: will using sql instead
	// init store client
	ing.boltClient = bolt.NewClient()
	ing.boltClient.Path = ing.boltPath
	ing.boltClient.WithLogger(ing.Logger.With(zap.String("service", "bbolt")))

	// Open bbolt
	if err := ing.boltClient.Open(ctx); err != nil {
		ing.Logger.Error("failed open", zap.String("service", "bbolt"), zap.Error(err))
		return err
	}

	serviceConfig := store.ServiceConfig{
		SessionLength: time.Duration(ing.sessionLength) * time.Minute,
	}

	// config store
	switch ing.storeType {
	case store.BblotStore:
		s := bolt.NewKVStore(ing.boltPath)
		s.WithDB(ing.boltClient.DB())
		ing.storeService = store.NewService(s, serviceConfig)
		// TODO: how to testing
	default:
		err := fmt.Errorf("unknown store type %s: excepted bolt", ing.storeType)
		ing.Logger.Error("expected bolt, unknown type", zap.String("store", ing.storeType))
		return err
	}

	// config log
	ing.storeService.Logger = ing.Logger.With(zap.String("store", ing.storeType))
	// init store
	if err := ing.storeService.Init(ctx); err != nil {
		ing.Logger.Error("failed to init store", zap.Error(err))
		return err
	}
	// define cache type
	// Now we config and init store in parse step.

	// define store type
	// registry metrics collector: GoCollector, serviceCollector
	ing.register = metrics.NewRegistry()
	ing.register.MustRegister(
		prometheus.NewGoCollector(),
		metrics.NewIndagateCollector("indagate", info),
	)
	ing.register.WithLogger(ing.Logger)
	// TODO: serviceCollector
	ing.register.MustRegister(ing.boltClient)
	// TODO: add other services
	var (
		auth service.AuthorizationService = ing.storeService
	)
	// todo： supported other store
	switch ing.secretType {
	case "bolt":
		// If bolt which has construct above
	case "valut":
		// TODO
	default:
		err := fmt.Errorf("unknown secret store type %s: excepted bolt", ing.secretType)
		ing.Logger.Error("expected bolt, vault ,unknown type", zap.String("store", ing.secretType))
		return err
	}
	// nats streaming for notify
	ing.natsServer = nats.NewServer()
	if err := ing.natsServer.Open(); err != nil {
		ing.Logger.Error("failed to start nats streaming server", zap.Error(err))
		return err
	}
	publisher := nats.NewAsyncPublisher("nats-publisher")
	// test open
	if err := publisher.Open(); err != nil {
		ing.Logger.Error("failed to connect to nats stream server", zap.Error(err))
		return err
	}
	ing.backend = &http.APIBackend{
		Logger: ing.Logger,
	}

	// registry prometheus metrics

	// build backend
	ing.backend = &http.APIBackend{
		Logger:                ing.Logger,
		AuthenticationService: auth,
	}

	// http logger
	httpLogger := ing.Logger.With(zap.String("service", "http"))
	platformHandler := routes.NewPlatformHandler(ing.backend)
	ing.register.MustRegister(platformHandler.PrometheusCollector()...)

	handler := http.NewHandlerWithRegistry("platform", ing.register)
	handler.Handler = platformHandler
	handler.Logger = httpLogger

	ing.server.Handler = handler
	// TODO: consider how to test server

	ls, err := net.Listen("tcp", ing.httpAddress.addr)
	if err != nil {
		httpLogger.Error("failed to listener ", zap.String("port", ing.httpAddress.addr), zap.Error(err))
		httpLogger.Info("Stoppping")
		return err
	}

	if addr, ok := ls.Addr().(*net.TCPAddr); ok {
		ing.httpAddress.port = addr.Port
	}

	ing.wg.Add(1)
	go func(logger *zap.Logger) {
		defer ing.wg.Done()

		logger.Info("Listening", zap.String("transport", "http"), zap.String("addr", ing.httpAddress.addr), zap.Int("port", ing.httpAddress.port))
		if err := server.LinstenAndServe(ing.httpAddress.addr, ing.server.Handler, logger); err != nil {
			logger.Error("failed start http service", zap.Error(err))
		}
		logger.Info("Stoppping")

	}(httpLogger)
	return nil
}
