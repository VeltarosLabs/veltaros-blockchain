package blockchain

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func (bc *Blockchain) persistPath(dataDir string) string {
	return filepath.Join(dataDir, "chain.json")
}

func (bc *Blockchain) SaveToDisk(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	path := bc.persistPath(dataDir)

	b, err := json.MarshalIndent(bc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func LoadFromDisk(dataDir string) (*Blockchain, error) {
	path := filepath.Join(dataDir, "chain.json")

	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var bc Blockchain
	if err := json.Unmarshal(b, &bc); err != nil {
		return nil, err
	}

	// NEW: rebuild UTXO set from blocks (Bitcoin style)
	// We attach it if your Blockchain struct has a field, otherwise you can keep it local for now.
	// If you already have bc.State for account balances, keep it, but don't require ApplyTransaction.
	//
	// If you add: bc.UTXO = NewUTXOSet() to Blockchain later, you can store it there.
	utxo := NewUTXOSet()
	utxo.Rebuild(bc.Blocks)

	// Optional: if Blockchain has a field for it, assign it:
	// bc.UTXO = utxo

	_ = utxo // keeps compile safe if you haven't added bc.UTXO yet

	return &bc, nil
}
