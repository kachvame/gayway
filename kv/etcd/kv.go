package etcd

import (
	"context"
	"fmt"
	"github.com/kachvame/gayway/kv"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type Store struct {
	client *clientv3.Client
}

func NewStore(endpoints []string, username, password string) (*Store, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		Username:    username,
		Password:    password,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial etcd: %w", err)
	}

	return &Store{client: client}, nil
}

func (store *Store) Get(ctx context.Context, key string) ([]byte, error) {
	response, err := store.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(response.Kvs) == 0 {
		return nil, kv.ErrNotFound
	}

	return response.Kvs[0].Value, nil
}

func (store *Store) Set(ctx context.Context, key string, value []byte) error {
	_, err := store.client.Put(ctx, key, string(value))

	return err
}

func (store *Store) Close() error {
	return store.client.Close()
}
