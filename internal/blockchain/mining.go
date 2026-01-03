package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/VeltarosLabs/veltaros-blockchain/pkg/crypto"
)

const Difficulty = 4

func MineBlock(block *Block) {
	for {
		hash := CalculateBlockHash(block)
		if strings.HasPrefix(hash, strings.Repeat("0", Difficulty)) {
			block.Hash = hash
			return
		}
		block.Nonce++
	}
}

// CalculateBlockHash MUST include transactions to prevent tampering.
func CalculateBlockHash(block *Block) string {
	txDigest := TransactionsDigest(block.Transactions)

	record := strconv.Itoa(block.Index) +
		strconv.FormatInt(block.Timestamp, 10) +
		txDigest +
		block.PrevHash +
		strconv.Itoa(block.Nonce)

	return crypto.GenerateHash(record)
}

// TransactionsDigest creates a deterministic digest of txs.
func TransactionsDigest(txs []Transaction) string {
	// Use JSON + SHA256 for deterministic digest (simple + stable).
	b, _ := json.Marshal(txs)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func IsPoWValid(hash string) bool {
	return strings.HasPrefix(hash, strings.Repeat("0", Difficulty))
}
