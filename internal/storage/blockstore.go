package storage

import (
	"encoding/json"
	"os"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
)

// BlockStore is a tiny JSON file store for the whole chain.
type BlockStore struct {
	Path string
}

func NewBlockStore(path string) *BlockStore {
	return &BlockStore{Path: path}
}

func (s *BlockStore) Save(chain []blockchain.Block) error {
	b, err := json.MarshalIndent(chain, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, b, 0o644)
}

func (s *BlockStore) Load() ([]blockchain.Block, error) {
	_, err := os.Stat(s.Path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, err
	}

	var chain []blockchain.Block
	if err := json.Unmarshal(b, &chain); err != nil {
		return nil, err
	}
	return chain, nil
}
