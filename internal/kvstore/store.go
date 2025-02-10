package kvstore

import (
	"bufio"
	"fmt"
	"os"

	// "path/filepath"
	"sync"
	kmap "github.com/jscottransom/distributed_godis/internal/keymap"
)

const (
	STORE_TEMPLATE = "godis_kv"
)

// Record is a struct representing a key value pairing
// Key must be a string and value must be a slice of bytes
type Record struct {
	Key   string
	Value []byte
}

type KVstore struct {
	file                   *os.File // File to work with
	Keymap		    	   *kmap.SafeMap
	mu                     sync.Mutex
	baseoffset, nextoffset uint64 // Represents the last offset in the file
	buf                    *bufio.Writer
}

func NewKVstore(dir string, name string) (*KVstore, error) {

	// Create new if it doesn't exist
	// err := os.MkdirAll(dir, os.ModePerm)
	// if err != nil {
	// 	return nil, fmt.Errorf("error creating directories %w", err)
	// }

	fileString := dir + "/" + name
	storefile, err := os.Create(fileString)
	if err != nil {
		return nil, fmt.Errorf("error opening file for writing to store: %w", err)
	}


	stat, err := os.Stat(storefile.Name())
	if err != nil {
		return nil, fmt.Errorf("error getting file stats: %w", err)
	}

	// Get the current offset of the file by getting the size
	offset := uint64(stat.Size())

	kmapObj, err := kmap.NewMap(dir)

	return &KVstore{file: storefile,
		Keymap: 	kmapObj,
		baseoffset: offset,
		nextoffset: offset,
		mu:         sync.Mutex{},
		buf:        bufio.NewWriter(storefile)}, nil
}

// Set the passed Key / Value pairing
func (s *KVstore) Set(record Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set the current offset to the value of the last offset in the store
	currentoffset := s.nextoffset

	key, err := s.buf.WriteString(record.Key)
	if err != nil {
		return err
	}

	// Increment the current offset by the number of bytes written to store the key
	currentoffset += uint64(key)

	// Write the values as bytes to the buffer
	value, err := s.buf.Write(record.Value)
	if err != nil {
		return err
	}

	// Increment the current offset by the number of bytes written to store the value
	currentoffset += uint64(value)

	// If we have successful writes to the buffer for key and value, update the store values
	s.nextoffset = currentoffset

	// Get the number of bytes for the value
	valueLen := uint64(len(record.Value))

	// Update the key in the keymap, and save the map
	keyinfo := kmap.KeyInfo{Size: valueLen,
		Offset: uint64(s.nextoffset) - valueLen}
	s.Keymap.FileLock.Lock()
	defer s.Keymap.FileLock.Unlock()
	s.Keymap.Map[record.Key] = &keyinfo
	s.Keymap.SaveMap()

	return nil

}

// Get the value for the specified key from the store
// Offset is the offset in the store to read
// N is the number of bytes to read
func (s *KVstore) Get(key string) ([]byte, error) {
	// Lock the file for safe access
	s.mu.Lock()
	defer s.mu.Unlock()

	// Flush any pending writes to disk
	s.buf.Flush()


	s.Keymap.FileLock.RLock()
	defer s.Keymap.FileLock.RUnlock()
	s.Keymap.LoadMap()

	keyInfo := s.Keymap.Map[key]
	value := make([]byte, keyInfo.Size)
	
	// Read the file at the given offset for the specified number of bytes
	val, err := s.file.ReadAt(value, int64(keyInfo.Offset))
	if err != nil {
		return nil, err
	}

	// Validate size of bytes read aligns with expectation
	if val != int(keyInfo.Size) {
		return nil, fmt.Errorf("error retrieving record due to incorrect byte size")
	}

	return value, nil

}

func (s *KVstore) Remove(dir string) error {
	
	return os.RemoveAll(dir)
}
