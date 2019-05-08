package store

import (
	"context"
	"errors"
)

const (
	BblotStore string = "bblot"
	MysqlStore string = "mysql"
)

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrTxNotWritable = errors.New("transaction is not writebale")
)

func IsNotFound(err error) bool {
	return err == ErrKeyNotFound
}

// Store define generic store interface
type Store interface {
	View(context.Context, func(Impl) error) error
	Modify(context.Context, func(Impl) error) error
}

// Impl is a transaction
type Impl interface {
	Bucket(b []byte) (Bucket, error)
	Context() context.Context
	WithContext(ctx context.Context)
}

// Bucket operation define
type Bucket interface {
	Get(key []byte) ([]byte, error)
	Cursor() (Cursor, error)
	Put(key, value []byte) error
	Delete(key []byte) error
}

// Cursor data selector
type Cursor interface {
	Seek(prefix []byte) (k, v []byte)
	First() (k, v []byte)
	Last() (k, v []byte)
	Next() (k, v []byte)
	Prev() (k, v []byte)
}
