package keymap

import (
	"encoding/gob"
	"fmt"
	"os"

	// "path/filepath"
	"sync"
)

const (
	MAP_TEMPLATE = "godis_keymap"
)

type KeyInfo struct {
	Size   uint64
	Offset uint64
}

// Simple abstraction to manage key lookups
// Deployed as an in-memory hash map (via go map)
type KeyMap map[string]*KeyInfo

type SafeMap struct {
	File  		*os.File
	FileLock  sync.RWMutex
	Map   	  KeyMap
}


// Instantiate a New SafeMap based on the directory of the KV Store
func NewMap(dir string) (*SafeMap, error) {
	// Create new if it doesn't exist
	mapString := dir + "/" + "keymap"
	mapFile, err := os.Create(mapString)
	if err != nil {
		fmt.Errorf("error opening file for writing Keymap: %w", err)
		return nil, err
	}

	mapobj := make(KeyMap, 0)

	return &SafeMap{
		File: mapFile,
		FileLock: sync.RWMutex{},
		Map: mapobj,
	}, nil


} 

// The following methods
// As the KeyMap is Updated, the map will be saved to a file
func (k *SafeMap) SaveMap() error {

	defer k.File.Close()
	// Instantiate a new Gob Encoder
	enc := gob.NewEncoder(k.File)
	err := enc.Encode(k.Map)
	if err != nil {
		return fmt.Errorf("error saving KeyMap: %w", err)
	}

	return nil

}

func (k *SafeMap) LoadMap() error {

	defer k.File.Close()

	// Instantiate a new Gob Encoder
	enc := gob.NewDecoder(k.File)
	if err := enc.Decode(&k.Map); err != nil {
		return fmt.Errorf("error loading KeyMap: %w", err)
	}

	return nil

}

func (k *SafeMap) SaveMap2(dir string, uid uint64) error {

	// Create new if it doesn't exist
	mapfile, err := os.Create("/Users/jscoran/godis/godis_keymap2")
	if err != nil {
		return fmt.Errorf("error creating file for writing to keymap: %w", err)
	}

	defer mapfile.Close()

	// Instantiate a new Gob Encoder
	enc := gob.NewEncoder(mapfile)
	err = enc.Encode(k.Map)
	if err != nil {
		return fmt.Errorf("error saving KeyMap: %w", err)
	}

	return nil

}
