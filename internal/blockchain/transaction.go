package blockchain

import (
	"strconv"

	"github.com/VeltarosLabs/veltaros-blockchain/pkg/crypto"
)

type Transaction struct {
	ID        string
	From      string
	To        string
	Amount    int
	Timestamp int64
}

func NewTransaction(from, to string, amount int) Transaction {
	tx := Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: now(),
	}
	tx.ID = txDigest(tx)
	return tx
}

// NewCoinbaseTransaction creates the mining reward tx (from "" to miner).
func NewCoinbaseTransaction(minerAddr string, amount int) Transaction {
	tx := Transaction{
		From:      "",
		To:        minerAddr,
		Amount:    amount,
		Timestamp: now(),
	}
	tx.ID = txDigest(tx)
	return tx
}

func txDigest(tx Transaction) string {
	return crypto.GenerateHash(
		tx.From + "|" +
			tx.To + "|" +
			strconv.Itoa(tx.Amount) + "|" +
			strconv.FormatInt(tx.Timestamp, 10),
	)
}
