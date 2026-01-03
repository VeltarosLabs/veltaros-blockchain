package blockchain

import "time"

// Blockchain represents the full chain and its mempool
type Blockchain struct {
	Blocks  []Block
	Mempool *Mempool
}

// NewBlockchain initializes the blockchain with genesis block
func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks:  []Block{GenesisBlock()},
		Mempool: NewMempool(),
	}
}

// AddTransaction adds a transaction to the mempool
func (bc *Blockchain) AddTransaction(tx Transaction) {
	bc.Mempool.AddTransaction(tx)
}

// MinePendingTransactions creates a new block from mempool txs
func (bc *Blockchain) MinePendingTransactions() {
	lastBlock := bc.Blocks[len(bc.Blocks)-1]

	newBlock := Block{
		Index:        lastBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: bc.Mempool.Flush(),
		PrevHash:     lastBlock.Hash,
	}

	MineBlock(&newBlock)
	bc.Blocks = append(bc.Blocks, newBlock)
}
