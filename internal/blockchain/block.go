package blockchain

import (
	"github.com/VeltarosLabs/veltaros-blockchain/internal/transaction"
)

type Block struct {
	Index        int
	Timestamp    int64
	Transactions []transaction.Transaction
	PrevHash     string
	Hash         string
	Nonce        int
}
