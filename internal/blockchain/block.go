package blockchain

import "time"

type Block struct {
	Index     int
	Timestamp int64
	Data      string
	PrevHash  string
	Hash      string
	Nonce     int
}

func NewBlock(index int, data string, prevHash string) *Block {
	return &Block{
		Index:     index,
		Timestamp: time.Now().Unix(),
		Data:      data,
		PrevHash:  prevHash,
	}
}
