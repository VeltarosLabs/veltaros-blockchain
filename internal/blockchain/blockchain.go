package blockchain

import (
	"time"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/transaction"
)

type Blockchain struct {
	Chain []Block
}

func NewBlockchain() *Blockchain {
	genesis := CreateGenesisBlock()
	return &Blockchain{
		Chain: []Block{genesis},
	}
}

func (bc *Blockchain) AddBlock(transactions []transaction.Transaction) {
	prevBlock := bc.Chain[len(bc.Chain)-1]

	newBlock := Block{
		Index:        prevBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PrevHash:     prevBlock.Hash,
	}

	MineBlock(&newBlock)

	if IsBlockValid(newBlock, prevBlock) {
		bc.Chain = append(bc.Chain, newBlock)
	}
}
