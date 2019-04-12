package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/ustackq/indagate/pkg/logger"
	"go.uber.org/zap"
)

// DefaultShutdownTimeout is the default timeout for shutting down the indagate server.
const DefaultShutdownTimeout = 20 * time.Second

// Server is an abstraction surround the http.Server which handles a server process.
type Server struct {
	ShutdownTimeout time.Duration
	srv             *http.Server
	signals         map[os.Signal]struct{}
	logger          *zap.Logger
	wg              sync.WaitGroup
}

// NewServer return a new server instance.
func NewServer(h http.Handler, logger *zap.Logger) *Server {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Server{
		ShutdownTimeout: DefaultShutdownTimeout,
		srv: &http.Server{
			Handler:  h,
			ErrorLog: zap.NewStdLog(logger),
		},
		logger: logger,
	}
}

func (s *Server) Serve(l net.Listener) error {
	defer s.wg.Wait()
	sigCh, cancel := s.notifyOnSignals()
	defer cancel()

	errCh := s.serve(l)
	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		return s.shutdown(sigCh)
	}
	return nil
}

func (s *Server) serve(l net.Listener) <-chan error {
	s.wg.Add(1)
	errCh := make(chan error, 1)
	go func() {
		defer s.wg.Done()
		if err := s.srv.Serve(l); err != nil {
			errCh <- err
		}
		close(errCh)
	}()
	return errCh
}

func (s *Server) shutdown(sigCh <-chan os.Signal) error {
	s.logger.Info("Shutting down server", logger.DurationLiteral("timeout", s.ShutdownTimeout))
	ctx, cancel := context.WithTimeout(context.Background(), s.ShutdownTimeout)
	defer cancel()

	done := make(chan struct{})
	defer close(done)
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		select {
		case <-sigCh:
			s.logger.Info("Hard shutdown.")
			cancel()
		case <-done:
		}
	}()
	return s.srv.Shutdown(ctx)
}

func (s *Server) ListenSignals(signals ...os.Signal) {
	if signals == nil {
		s.signals = make(map[os.Signal]struct{})
	}
	for _, sig := range signals {
		s.signals[sig] = struct{}{}
	}
}

func (s *Server) notifyOnSignals() (_ <-chan os.Signal, cancel func()) {
	if len(s.signals) == 0 {
		return nil, func() {}
	}
	signals := make([]os.Signal, 0, len(s.signals))
	for sig := range s.signals {
		signals = append(signals, sig)
	}

	sigCh := make(chan os.Signal, 2*len(signals))
	signal.Notify(sigCh, signals...)
	return sigCh, func() { signal.Stop(sigCh) }
}
