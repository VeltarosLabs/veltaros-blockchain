package blockchain

import "time"

// Blockchain represents the full chain and its mempool
type Blockchain struct {
	Blocks  []Block
	Mempool *Mempool
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks:  []Block{GenesisBlock()},
		Mempool: NewMempool(),
	}
}

func (bc *Blockchain) AddTransaction(tx Transaction) error {
	return bc.Mempool.AddTransaction(tx)
}

func (bc *Blockchain) MinePendingTransactions() Block {
	last := bc.Blocks[len(bc.Blocks)-1]

	newBlock := Block{
		Index:        last.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: bc.Mempool.Flush(),
		PrevHash:     last.Hash,
	}

	MineBlock(&newBlock)
	bc.Blocks = append(bc.Blocks, newBlock)
	return newBlock
}

// TryAddBlock validates then appends a network block.
func (bc *Blockchain) TryAddBlock(b Block) bool {
	last := bc.Blocks[len(bc.Blocks)-1]
	if !IsBlockValid(b, last) {
		return false
	}
	bc.Blocks = append(bc.Blocks, b)
	return true
}

// TryReplaceChain applies fork rule: prefer longer valid chain.
func (bc *Blockchain) TryReplaceChain(candidate []Block) bool {
	if len(candidate) <= len(bc.Blocks) {
		return false
	}
	if !IsChainValid(candidate) {
		return false
	}
	bc.Blocks = candidate
	return true
}
