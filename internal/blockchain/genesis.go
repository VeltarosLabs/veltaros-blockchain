package blockchain

import "time"

func CreateGenesisBlock() Block {
	block := Block{
		Index:     0,
		Timestamp: time.Now().Unix(),
		PrevHash:  "0",
	}
	MineBlock(&block)
	return block
}
