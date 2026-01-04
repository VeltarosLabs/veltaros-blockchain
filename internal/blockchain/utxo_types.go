package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// OutPoint identifies a specific output of a previous transaction.
type OutPoint struct {
	TxID string `json:"txid"`
	Vout int    `json:"vout"`
}

type TxIn struct {
	PrevOut OutPoint `json:"prev_out"`
	// For now we keep signatures simple; we'll add real ECDSA script later.
	Signature []byte `json:"sig,omitempty"`
	PubKey    []byte `json:"pubkey,omitempty"`
}

type TxOut struct {
	Value      int    `json:"value"`
	PubKeyHash []byte `json:"pubkey_hash"` // lock to address-hash
}

type UTXOTransaction struct {
	ID        string  `json:"id"`
	Vin       []TxIn  `json:"vin"`
	Vout      []TxOut `json:"vout"`
	Timestamp int64   `json:"timestamp"`
}

func (tx *UTXOTransaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && tx.Vin[0].PrevOut.TxID == "" && tx.Vin[0].PrevOut.Vout == -1
}

func (tx *UTXOTransaction) Hash() string {
	clone := *tx
	clone.ID = ""
	b, _ := json.Marshal(clone)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func NewCoinbaseUTXOTx(toPubKeyHash []byte, reward int) UTXOTransaction {
	tx := UTXOTransaction{
		Vin: []TxIn{
			{PrevOut: OutPoint{TxID: "", Vout: -1}},
		},
		Vout: []TxOut{
			{Value: reward, PubKeyHash: toPubKeyHash},
		},
		Timestamp: time.Now().Unix(),
	}
	tx.ID = tx.Hash()
	return tx
}

// Helper: compare pubkeyhash
func samePubKeyHash(a, b []byte) bool { return bytes.Equal(a, b) }

// Debug / safety
func (op OutPoint) String() string {
	return fmt.Sprintf("%s:%d", op.TxID, op.Vout)
}
