package transaction

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Transaction struct {
	ID        string
	From      string
	To        string
	Amount    float64
	Timestamp int64
}

func NewTransaction(from, to string, amount float64) Transaction {
	tx := Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
	}
	tx.ID = tx.calculateHash()
	return tx
}

func (tx *Transaction) calculateHash() string {
	record := tx.From + tx.To
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}
