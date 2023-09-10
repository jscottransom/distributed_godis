package keymap

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

const (
	MAP_TEMPLATE = "godis_keymap_%d.map"
)

type KeyInfo struct {
	size   uint64
	offset uint64
}

// Simple abstraction to manage key lookups
// Deployed as an in-memory hash map (via go map)
type KeyMap map[string]KeyInfo

// The following methods
// As the KeyMap is Updated, the map will be saved to a file
func (k KeyMap) SaveMap(dir string, uid uint64) error {

	// Create new if it doesn't exist
	path := filepath.Join(dir, fmt.Sprintf(MAP_TEMPLATE, uid))
	mapfile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file for writing to keymap: %w", err)
	}

	defer mapfile.Close()

	// Instantiate a new Gob Encoder
	enc := gob.NewEncoder(mapfile)
	if err := enc.Encode(k); err != nil {
		return fmt.Errorf("error saving KeyMap: %w", err)
	}

	return nil

}

func (k KeyMap) LoadMap(dir string, uid uint64) error {

	// Create new if it doesn't exist
	path := filepath.Join(dir, fmt.Sprintf(MAP_TEMPLATE, uid))
	mapfile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file for writing to store: %w", err)
	}

	defer mapfile.Close()

	// Instantiate a new Gob Encoder
	enc := gob.NewDecoder(mapfile)
	if err := enc.Decode(&k); err != nil {
		return fmt.Errorf("error loading KeyMap: %w", err)
	}

	return nil

}
