package bolt

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"

	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/generator"
	bolt "go.etcd.io/bbolt"
)

// OpPrefix is the prefix for bolt ops
const OpPrefix = "bolt/"

func getOp(op string) string {
	return OpPrefix + op
}

type Client struct {
	Path   string
	db     *bolt.DB
	Logger *zap.Logger

	IDGenerator    service.IDGenerator
	TokenGenerator generator.TokenGenerator
	time           func() time.Time
}

func NewClient() *Client {
	return &Client{
		time:           time.Now,
		Logger:         zap.NewNop(),
		IDGenerator:    generator.NewIDGenerator(),
		TokenGenerator: generator.NewTokenGenerator(64),
	}
}

func (client *Client) DB() *bolt.DB {
	return client.db
}

func (client *Client) WithTime(f func() time.Time) {
	client.time = f
}

// Open ensure file exist
func (client *Client) Open(ctx context.Context) error {
	if err := os.MkdirAll(filepath.Dir(client.Path), 0700); err != nil {
		return fmt.Errorf("unable to create dir %s: %v", client.Path, err)
	}

	if _, err := os.Stat(client.Path); err != nil && !os.IsNotExist(err) {
		return err
	}

	db, err := bolt.Open(client.Path, 0700, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return fmt.Errorf("unable to open boltdb: %v", err)
	}
	client.db = db
	if err := client.init(ctx); err != nil {
		return err
	}

	client.Logger.Info("Resource boltdb opened", zap.String("path", client.Path))
	return nil
}

func (client *Client) Close() error {
	if client.db != nil {
		client.Logger.Info("Resource boltdb closing", zap.String("path", client.Path))
		return client.db.Close()
	}
	client.Logger.Warn("Resource boltdb is nil", zap.String("path", client.Path))
	return nil
}

func (client *Client) init(ctx context.Context) error {
	return nil
}

func (client *Client) WithLogger(logger *zap.Logger) {
	client.Logger = logger
}
