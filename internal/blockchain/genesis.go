package blockchain

import (
	"fmt"

	"github.com/VeltarosLabs/veltaros-blockchain/pkg/crypto"
)

func CreateGenesisBlock() *Block {
	block := NewBlock(0, "Genesis Block", "0")
	block.Hash = crypto.CalculateHash(fmt.Sprintf(
		"%d%d%s%s",
		block.Index,
		block.Timestamp,
		block.Data,
		block.PrevHash,
	))
	return block
}
