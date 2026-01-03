package blockchain

// Block represents a single block
type Block struct {
	Index        int
	Timestamp    int64
	Transactions []Transaction
	PrevHash     string
	Hash         string
	Nonce        int
}
