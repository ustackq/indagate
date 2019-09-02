package bolt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"

	"github.com/ustackq/indagate/pkg/store"
	"github.com/ustackq/indagate/pkg/tracing"
)

type KVStore struct {
	path   string
	db     *bolt.DB
	logger *zap.Logger
}

func NewKVStore(path string) *KVStore {
	return &KVStore{
		path:   path,
		logger: zap.NewNop(),
	}
}

// Open create boltDB file if it doesn't exist, otherwise opens it.
func (kv *KVStore) Open(ctx context.Context) error {
	span, _ := tracing.StartSpanFromContext(ctx)
	defer span.End()

	if err := os.MkdirAll(filepath.Dir(kv.path), 0700); err != nil {
		return fmt.Errorf("unable to create directory %s: %v", kv.path, err)
	}

	if _, err := os.Stat(kv.path); err != nil {
		return err
	}

	db, err := bolt.Open(kv.path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return fmt.Errorf("unable open boltdb file %s: %v", kv.path, err)
	}

	kv.db = db

	kv.logger.Info("Resources opened", zap.String("path", kv.path))
	return nil
}

func (kv *KVStore) Close() error {
	if kv.db != nil {
		return kv.db.Close()
	}

	return nil
}

func (kv *KVStore) Flush(ctx context.Context) {
	kv.db.Update(
		func(tx *bolt.Tx) error {
			return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
				kv.cleanBucket(tx, b)
				return nil
			})
		},
	)
}

func (kv *KVStore) cleanBucket(tx *bolt.Tx, b *bolt.Bucket) {
	if b == nil {
		return
	}

	c := b.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		_ = v
		if err := c.Delete(); err != nil {
			kv.cleanBucket(tx, b.Bucket(k))
		}
	}
}

func (kv *KVStore) WithLogger(l *zap.Logger) {
	kv.logger = l
}

func (kv *KVStore) WithDB(db *bolt.DB) {
	kv.db = db
}

func (kv *KVStore) View(ctx context.Context, fn func(tx store.Impl) error) error {
	// add tracing
	span, ctx := tracing.StartSpanFromContext(ctx)
	defer span.End()

	return kv.db.View(
		func(tx *bolt.Tx) error {
			return fn(&Tx{
				tx:  tx,
				ctx: ctx,
			})
		},
	)
}

func (kv *KVStore) Modify(ctx context.Context, fn func(tx store.Impl) error) error {
	span, ctx := tracing.StartSpanFromContext(ctx)
	defer span.End()

	return kv.db.Update(
		func(tx *bolt.Tx) error {
			return fn(&Tx{
				tx:  tx,
				ctx: ctx,
			})
		},
	)
}

// Tx wrap boltdb transaction, It implements kv.Tx.
type Tx struct {
	tx  *bolt.Tx
	ctx context.Context
}

// Context returns the context for the transaction.
func (tx *Tx) Context() context.Context {
	return tx.ctx
}

// WithContext sets the context for the transaction.
func (tx *Tx) WithContext(ctx context.Context) {
	tx.ctx = ctx
}

// Bucket retrieves the bucket according args
func (tx *Tx) Bucket(b []byte) (store.Bucket, error) {
	bkt := tx.tx.Bucket(b)
	if bkt == nil {
		return tx.createBucketIfNotExists(b)
	}
	return &Bucket{
		bucket: bkt,
	}, nil
}

func (tx *Tx) createBucketIfNotExists(b []byte) (*Bucket, error) {
	bkt, err := tx.tx.CreateBucketIfNotExists(b)
	if err != nil {
		return nil, err
	}
	return &Bucket{
		bucket: bkt,
	}, nil
}

// Bucket implements kv.Bucket.
type Bucket struct {
	bucket *bolt.Bucket
}

// Get retrieves the value at the provided key.
func (b *Bucket) Get(key []byte) ([]byte, error) {
	val := b.bucket.Get(key)
	if len(val) == 0 {
		return nil, store.ErrKeyNotFound
	}

	return val, nil
}

// Put sets the value at the provided key.
func (b *Bucket) Put(key []byte, value []byte) error {
	err := b.bucket.Put(key, value)
	if err == bolt.ErrTxNotWritable {
		return store.ErrTxNotWritable
	}
	return err
}

// Delete removes the provided key.
func (b *Bucket) Delete(key []byte) error {
	err := b.bucket.Delete(key)
	if err == bolt.ErrTxNotWritable {
		return store.ErrTxNotWritable
	}
	return err
}

// Cursor retrieves a cursor for iterating through the entries
// in the key value store.
func (b *Bucket) Cursor() (store.Cursor, error) {
	return &Cursor{
		cursor: b.bucket.Cursor(),
	}, nil
}

// Cursor is a struct for iterating through the entries
// in the key value store.
type Cursor struct {
	cursor *bolt.Cursor
}

// Seek seeks for the first key that matches the prefix provided.
func (c *Cursor) Seek(prefix []byte) ([]byte, []byte) {
	k, v := c.cursor.Seek(prefix)
	if len(k) == 0 && len(v) == 0 {
		return nil, nil
	}
	return k, v
}

// First retrieves the first key value pair in the bucket.
func (c *Cursor) First() ([]byte, []byte) {
	k, v := c.cursor.First()
	if len(k) == 0 && len(v) == 0 {
		return nil, nil
	}
	return k, v
}

// Last retrieves the last key value pair in the bucket.
func (c *Cursor) Last() ([]byte, []byte) {
	k, v := c.cursor.Last()
	if len(k) == 0 && len(v) == 0 {
		return nil, nil
	}
	return k, v
}

// Next retrieves the next key in the bucket.
func (c *Cursor) Next() ([]byte, []byte) {
	k, v := c.cursor.Next()
	if len(k) == 0 && len(v) == 0 {
		return nil, nil
	}
	return k, v
}

// Prev retrieves the previous key in the bucket.
func (c *Cursor) Prev() ([]byte, []byte) {
	k, v := c.cursor.Prev()
	if len(k) == 0 && len(v) == 0 {
		return nil, nil
	}
	return k, v
}
