package kv

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("kv: not found")

type Store interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
	Close() error
}
