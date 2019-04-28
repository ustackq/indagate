package config

import "strings"

type Store interface {
	// Load updates the current configuration from the backing store, possibly initializing.
	Load() (err error)
	Flush()
}

func NewStore(cfg string, watch bool) (Store, error) {
	if strings.HasPrefix(cfg, "mysql://") || strings.HasPrefix(cfg, "postgres://") {
		return NewDatabaseStore(cfg)
	}
	return nil, nil
}
