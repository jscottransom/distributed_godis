package kvstore

import (
	"fmt"
	"sync"
	"os"
	"path/filepath"
	"bufio"

)

const (
	STORE_TEMPLATE = "godis_kv_d.kv%"
)

// Record is a struct representing a key value pairing
// Key must be a string and value must be a slice of bytes
type Record struct {
	key string
	value []byte
}

type KVstore struct {
	file *os.File // File to work with
	mu sync.Mutex
	baseoffset, nextoffset uint64 // Represents the last offset in the file
	buf *bufio.Writer 

}

func NewKVstore(dir string, uid uint64) (*KVstore, error) {

	// Create new if it doesn't exist
	path := filepath.Join(dir, fmt.Sprintf(STORE_TEMPLATE, uid))
	storefile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening file for writing to store: %w", err)
	}

	stat, err := os.Stat(storefile.Name())

	// Get the current offset of the file by getting the size
	offset := uint64(stat.Size())

	return &KVstore{file: storefile,
					baseoffset: offset,
					nextoffset: offset,
					mu: sync.Mutex{},
					buf: bufio.NewWriter(storefile)}, nil
}

// Set the passed Key / Value pairing
func (s *KVstore) set(record Record) (uint64, error) {
	fmt.Printf("Setting the following values: %s, %s", record.key, string(record.value))
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set the current offset to the value of the last offset in the store
	currentoffset := s.nextoffset

	key, err := s.buf.WriteString(record.key)
	if err != nil {
		return 0, err
	}

	// Increment the current offset by the number of bytes written to store the key
	currentoffset += uint64(key)

	// Write the values as bytes to the buffer 
	value, err := s.buf.Write(record.value)
	if err != nil {
		return 0, err
	}

	// Increment the current offset by the number of bytes written to store the value
	currentoffset += uint64(value)

	// If we have successful writes to the buffer for key and value, update the store values
	s.nextoffset = currentoffset

	return currentoffset, nil

	

		
	

}

// Get the value for the specified key from the store
// Offset is the offset in the store to read
// N is the number of bytes to read
func (s *KVstore) get(offset uint64, n uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Flush any pending writes to disk
	s.buf.Flush()

	value := make([]byte, n)
	// Read the file at the given offset for the specified number of bytes
	val, err := s.file.ReadAt(value, int64(offset))
	if err != nil {
		return nil, err
	}

	// Validate size of bytes read aligns with expectation
	if val != int(n) {
		return nil, fmt.Errorf("Error retrieving record due to incorrect byte size.")
	}

	return value, nil

}
