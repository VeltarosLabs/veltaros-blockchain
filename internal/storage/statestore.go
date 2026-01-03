package storage

import (
	"encoding/json"

	bolt "go.etcd.io/bbolt"
)

const StateBucket = "state"

// SaveState stores the whole state map under key "balances".
func SaveState(db *bolt.DB, state map[string]int) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(StateBucket))
		if err != nil {
			return err
		}
		return b.Put([]byte("balances"), data)
	})
}

// LoadState loads the whole state map from key "balances".
func LoadState(db *bolt.DB) (map[string]int, error) {
	state := make(map[string]int)

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(StateBucket))
		if b == nil {
			return nil
		}
		data := b.Get([]byte("balances"))
		if data == nil {
			return nil
		}
		return json.Unmarshal(data, &state)
	})

	return state, err
}
