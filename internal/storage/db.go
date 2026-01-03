package storage

import (
	"encoding/json"
	"os"
	"sync"
)

// Simple file-backed key/value store (no blockchain imports => no cycles).
type DB struct {
	mu   sync.Mutex
	path string
	data map[string][]byte
}

func Open(path string) (*DB, error) {
	db := &DB{
		path: path,
		data: map[string][]byte{},
	}

	// Load if exists
	if b, err := os.ReadFile(path); err == nil && len(b) > 0 {
		_ = json.Unmarshal(b, &db.data)
	}

	return db, nil
}

func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	b, err := json.MarshalIndent(db.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(db.path, b, 0644)
}

func (db *DB) Put(key string, val []byte) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.data[key] = val
}

func (db *DB) Get(key string) ([]byte, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()
	v, ok := db.data[key]
	return v, ok
}
