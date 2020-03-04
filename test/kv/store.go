package kv

import (
	"fmt"
	"log"
	"os"
)

type Store struct {
	store map[string]string
}

func NewStore() *Store {
	return &Store{
		store: map[string]string{},
	}
}

type SetArg struct {
	Key string
	Value string
}

type GetArg struct {
	Key string
}

func (s *Store) Set(arg SetArg, _ *int) error {
	log.Println("#Set")

	s.store[arg.Key] = arg.Value
	return nil
}

func (s *Store) Get(arg GetArg, reply *string) error {
	log.Println("#Get")

	val, ok := s.store[arg.Key]
	if !ok {
		return fmt.Errorf("key '%s' not present", arg.Key)
	}

	*reply = val
	return nil
}

func (s *Store) List(arg int, reply *[]string) error {
	log.Println("#List")

	var keys []string
	for key := range s.store {
		keys = append(keys, key)
	}

	*reply = keys
	return nil
}

func (s *Store) Destroy(arg int, _ *int) error {
	log.Println("#Destroy")
	os.Exit(0)
	return nil
}
