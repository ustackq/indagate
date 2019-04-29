package nats

import (
	"errors"
	stream "github.com/nats-io/go-nats-streaming"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
)

const NatsServer = "Indagate"

var ErrNatsConnection = errors.New("nats connection has not been established.")

type Publisher interface {
	Publish(subject string, reader io.Reader) error
}

type AsyncPubliber struct {
	ClientID string
	Conn     stream.Conn
	Logger   *zap.Logger
}

func NewAsyncPublisher(id string) *AsyncPubliber {
	return &AsyncPubliber{ClientID: id}
}

func (ap *AsyncPubliber) Publish(subject string, reader io.Reader) error {
	if ap.Conn == nil {
		return ErrNatsConnection
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	ac := func(s string, err error) {
		if err != nil {
			ap.Logger.Info(err.Error())
		}
	}
	_, err = ap.Conn.PublishAsync(subject, data, ac)
	return err
}

func (ap *AsyncPubliber) Open() error {
	conn, err := stream.Connect(NatsServer, ap.ClientID)
	if err != nil {
		return err
	}
	ap.Conn = conn
	return nil
}
