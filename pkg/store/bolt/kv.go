package bolt

import (
	"context"

	"github.com/coreos/bbolt"
	"go.uber.org/zap"

	"github.com/ustackq/indagate/pkg/store"
)

type KVStore struct {
	path   string
	db     *bbolt.DB
	logger *zap.Logger
}

func NewKVStore(path string) *KVStore {
	return &KVStore{
		path:   path,
		logger: zap.NewNop(),
	}
}

func (kv *KVStore) WithDB(db *bbolt.DB) {
	kv.db = db
}

func (kv *KVStore) View(ctx context.Context, fn func(tx store.Impl) error) error {
	// TODO: add tracing
	return kv.db.View(
		func(tx *bbolt.Tx) error {
			return fn(&Tx{
				tx:  tx,
				ctx: ctx,
			})
		},
	)
}

func (kv *KVStore) Modify(ctx context.Context, fn func(tx store.Impl) error) error {
	// TODO: add tracing
	return kv.db.Update(
		func(tx *bbolt.Tx) error {
			return fn(&Tx{
				tx:  tx,
				ctx: ctx,
			})
		},
	)
}

type Tx struct {
	tx  *bbolt.Tx
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

func (tx *Tx) Bucket(b []byte) (store.Bucket, error) {
	bkt := tx.tx.Bucket(b)
	if bkt == nil {
		return tx.createBucketIfNotExists(b)
	}
	return &Bucket{
		bucket: bkt,
	}, nil
}

func (tx *Tx) createBucketIfNotExists(b []byte) (store.Bucket, error) {
	bkt, err := tx.tx.CreateBucketIfNotExists(b)
	if err != nil {
		return nil, err
	}
	return &Bucket{
		bucket: bkt,
	}, nil
}

type Bucket struct {
	bucket *bbolt.Bucket
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
	if err == bbolt.ErrTxNotWritable {
		return store.ErrTxNotWritable
	}
	return err
}

// Delete removes the provided key.
func (b *Bucket) Delete(key []byte) error {
	err := b.bucket.Delete(key)
	if err == bbolt.ErrTxNotWritable {
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
	cursor *bbolt.Cursor
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
