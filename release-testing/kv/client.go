package kv

import (
	"fmt"
	"net/rpc"
	"os"
)

type StoreClient struct {
	client *rpc.Client
}

func NewStoreClient(network string, address string) (*StoreClient, error) {
	client, err := rpc.DialHTTP(network, address)
	if err != nil {
		return nil, err
	}

	return &StoreClient{
		client: client,
	}, nil
}

func (s *StoreClient) Set(key, val string) error {
	args := SetArg{
		Key: key,
		Value: val,
	}

	return s.client.Call("Store.Set", args, nil)
}

func (s *StoreClient) Get(key string) (string, error) {
	var val string
	err := s.client.Call("Store.Get", GetArg{Key: key}, &val)
	if err != nil {
		return "", err
	}

	return val, nil
}

func (s *StoreClient) List() ([]string, error) {
	var keys []string
	err := s.client.Call("Store.List", 0, &keys)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (s *StoreClient) Destroy() error {
	return s.client.Call("Store.Destroy", 0, nil)
}

func DefaultStoreClient() (*StoreClient, error) {
	port, ok := os.LookupEnv("STORE_PORT")
	if !ok {
		return nil, fmt.Errorf("STORE_PORT envvar not set")
	}

	return NewStoreClient("tcp", "localhost:" + port)
}
