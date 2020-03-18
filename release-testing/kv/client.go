package kv

import (
	"fmt"
	"net/rpc"
	"os"
)

// StoreClient is an rpc client for a Store rpc server
type StoreClient struct {
	client *rpc.Client
}

// NewStoreClient creates an rpc client for a Store rpc server
func NewStoreClient(network string, address string) (*StoreClient, error) {
	client, err := rpc.DialHTTP(network, address)
	if err != nil {
		return nil, err
	}

	return &StoreClient{
		client: client,
	}, nil
}

// Set calls the Set method on the Store
func (s *StoreClient) Set(key, val string) error {
	args := SetArg{
		Key:   key,
		Value: val,
	}

	return s.client.Call("Store.Set", args, nil)
}

// Get calls the Get method on the Store
func (s *StoreClient) Get(key string) (string, error) {
	var val string
	err := s.client.Call("Store.Get", GetArg{Key: key}, &val)
	if err != nil {
		return "", err
	}

	return val, nil
}

// List calls the List method on the Store
func (s *StoreClient) List() ([]string, error) {
	var keys []string
	err := s.client.Call("Store.List", 0, &keys)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

// Snapshot calls the Snapshot method on the Store
func (s *StoreClient) Snapshot() (string, error) {
	var out string
	err := s.client.Call("Store.Snapshot", 0, &out)
	if err != nil {
		return "", err
	}

	return out, nil
}

// Destroy calls the Destroy method on the Store
func (s *StoreClient) Destroy() error {
	return s.client.Call("Store.Destroy", 0, nil)
}

// DefaultStoreClient creates an rpc client for a Store rpc server using default values
// of network=tcp, host=localhost, and port=the 'STORE_PORT' environment variable.
func DefaultStoreClient() (*StoreClient, error) {
	port, ok := os.LookupEnv("STORE_PORT")
	if !ok {
		return nil, fmt.Errorf("STORE_PORT envvar not set")
	}

	return NewStoreClient("tcp", "localhost:"+port)
}
