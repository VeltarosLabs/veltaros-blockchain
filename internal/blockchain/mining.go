package blockchain

import (
	"strconv"
	"strings"

	"github.com/VeltarosLabs/veltaros-blockchain/pkg/crypto"
)

const Difficulty = 4

func MineBlock(block *Block) {
	for {
		hash := calculateBlockHash(block)
		if strings.HasPrefix(hash, strings.Repeat("0", Difficulty)) {
			block.Hash = hash
			break
		}
		block.Nonce++
	}
}

func calculateBlockHash(block *Block) string {
	record := strconv.Itoa(block.Index) +
		strconv.FormatInt(block.Timestamp, 10) +
		block.PrevHash +
		strconv.Itoa(block.Nonce)

	return crypto.GenerateHash(record)
}
