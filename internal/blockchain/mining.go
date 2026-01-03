package blockchain

import (
	"strconv"
	"strings"
	"time"

	"github.com/VeltarosLabs/veltaros-blockchain/pkg/crypto"
)

// now returns unix timestamp (seconds).
func now() int64 {
	return time.Now().Unix()
}

// SimpleSHA256 is a small helper used by some code paths.
// It returns a hex string hash of input.
func SimpleSHA256(input string) string {
	return crypto.GenerateHash(input)
}

// CalculateBlockHash calculates the hash of a block.
func CalculateBlockHash(block *Block) string {
	record := strconv.Itoa(block.Index) +
		strconv.FormatInt(block.Timestamp, 10) +
		block.PrevHash +
		TransactionsDigest(block.Transactions) +
		strconv.Itoa(block.Nonce)

	return crypto.GenerateHash(record)
}

// IsPoWValid checks if a hash satisfies Difficulty leading zeros.
func IsPoWValid(hash string) bool {
	return strings.HasPrefix(hash, strings.Repeat("0", Difficulty))
}

// MineBlock increments nonce until PoW is valid and sets block.Hash.
func MineBlock(block *Block) {
	for {
		hash := CalculateBlockHash(block)
		if IsPoWValid(hash) {
			block.Hash = hash
			return
		}
		block.Nonce++
	}
}

// TransactionsDigest creates a deterministic digest of txs.
// It avoids relying on tx.ID (since your Transaction type doesn’t have ID).
func TransactionsDigest(txs []Transaction) string {
	var b strings.Builder
	for _, tx := range txs {
		// Keep it stable + simple:
		b.WriteString(tx.From)
		b.WriteString("->")
		b.WriteString(tx.To)
		b.WriteString(":")
		b.WriteString(strconv.Itoa(tx.Amount))
		b.WriteString("@")
		b.WriteString(strconv.FormatInt(tx.Timestamp, 10))
		b.WriteString("|")
	}
	return crypto.GenerateHash(b.String())
}
