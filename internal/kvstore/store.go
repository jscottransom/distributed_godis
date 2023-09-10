package kvstore

import (
	"fmt"
	"sync"
)

// KVstore is a map object that takes a key which must be a string and a value which can be of any type
type KV map[string]interface{}

type KVstore struct {
	path string
	store KV
	mu sync.Mutex // guards

}

func NewKVstore(path string) (*KVstore, error) {
	// Initialize an empty Map object
	kv := make(KV)
	
	return &KVstore{path: path,
					store: kv,
					mu: sync.Mutex{}}, nil
}

// Set the passed Key / Value pairing
// You can safely pass around a map
// Pass-by-value for maps actually updates the underlying
func (s *KVstore) Set(key string, value interface{}) error {
	fmt.Printf("Setting the following values: %s, %s", key, value)
	s.mu.Lock()
	defer s.mu.Unlock()
	// Check if the key already exists
	_, exists := s.store[key]
	if exists {
		fmt.Printf("Key already exists. Setting to: %s\n", value)
		s.store[key] = value
		return nil
	} else {
		fmt.Printf("Key doesn't exist. Setting to: %s\n", value)
		s.store[key] = value
		return nil
	}

}

// Get the value for the specified key from the store
func (s *KVstore) Get(key string) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Check if the key exists
	value, ok := s.store[key]
	if ok {
		fmt.Printf("Value for key: %s, is %s\n", key, value)
		return value, nil
	} else {
		return nil, nil
	}

}
