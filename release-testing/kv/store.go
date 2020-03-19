package kv

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cyberark/conjur-oss-suite-release/pkg/log"
)

// Store is an in-memory key-value store
type Store map[string]string

// SetArg holds the arguments for the Set method on the Store
type SetArg struct {
	Key   string
	Value string
}

// GetArg holds the arguments for the Get method on the Store
type GetArg struct {
	Key string
}

// Set inserts a key-value pair into the store
func (s Store) Set(arg SetArg, _ *int) error {
	log.OutLogger.Println("#Set")

	s[arg.Key] = arg.Value
	return nil
}

// Get retrieves a value from the store, given a key.
// It returns an error if the key does not exist
func (s Store) Get(arg GetArg, reply *string) error {
	log.OutLogger.Println("#Get")

	val, ok := s[arg.Key]
	if !ok {
		return fmt.Errorf("key '%s' not present", arg.Key)
	}

	*reply = val
	return nil
}

// List returns a slice of strings containing each of the keys present in the store
func (s Store) List(arg int, reply *[]string) error {
	log.OutLogger.Println("#List")

	var keys []string
	for key := range s {
		keys = append(keys, key)
	}

	*reply = keys
	return nil
}

// Snapshot returns a JSON serialized string of the Store at the moment the call is made
func (s Store) Snapshot(arg int, reply *string) error {
	log.OutLogger.Println("#Snapshot")

	res, err := json.Marshal(s)
	*reply = string(res)

	return err
}

// Destroy kills the process with 0 exit code
func (s Store) Destroy(arg int, _ *int) error {
	log.OutLogger.Println("#Destroy")
	os.Exit(0)
	return nil
}
