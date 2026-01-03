package blockchain

import (
	"fmt"

	"github.com/VeltarosLabs/veltaros-blockchain/pkg/crypto"
)

type Blockchain struct {
	Blocks []*Block
}

func NewBlockchain() *Blockchain {
	genesis := CreateGenesisBlock()
	return &Blockchain{
		Blocks: []*Block{genesis},
	}
}

func (bc *Blockchain) AddBlock(data string) {
	prev := bc.Blocks[len(bc.Blocks)-1]
	newBlock := NewBlock(len(bc.Blocks), data, prev.Hash)

	newBlock.Hash = crypto.CalculateHash(fmt.Sprintf(
		"%d%d%s%s",
		newBlock.Index,
		newBlock.Timestamp,
		newBlock.Data,
		newBlock.PrevHash,
	))

	bc.Blocks = append(bc.Blocks, newBlock)
}
