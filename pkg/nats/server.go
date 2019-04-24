package nats

import (
	"github.com/nats-io/nats-streaming-server/server"
)

// Server wraps a connection to a NATS streaming server
type Server struct {
	Server      *server.StanServer
	StoreConfig string
	ID          string
}

// NewServer return a new Server struct
func NewServer() *Server {
	return &Server{}
}

// Open starts a NASTs streaming server
func (s *Server) Open() error {
	stanOpts := server.GetDefaultOptions()
	stanOpts.StoreType = s.StoreConfig
	stanOpts.ID = s.ID
	// TODO: supported server.Option
	server, err := server.RunServerWithOpts(stanOpts, nil)
	if err != nil {
		return err
	}
	s.Server = server
	return nil
}

// Close stops NATS service
func (s *Server) Close() {
	s.Server.Shutdown()
}
