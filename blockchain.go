package blockchain

import "sync"

type Blockchain struct {
	mu      sync.Mutex
	Blocks  []Block
	Mempool *Mempool
	State   *State
}

func NewBlockchain() *Blockchain {
	bc := &Blockchain{
		Blocks:  []Block{GenesisBlock()},
		Mempool: NewMempool(),
		State:   NewState(),
	}

	// Apply genesis (no txs, but keeps logic consistent)
	_ = bc.State.ApplyBlock(bc.Blocks[0])

	return bc
}

// AddTransaction adds tx to mempool (minimal checks).
func (bc *Blockchain) AddTransaction(tx Transaction) error {
	if tx.Amount < 0 {
		return ErrInvalidTransaction
	}
	bc.Mempool.AddTransaction(tx)
	return nil
}

func (bc *Blockchain) MinePendingTransactions(minerAddr string) (Block, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if minerAddr == "" {
		return Block{}, ErrInvalidTransaction
	}

	last := bc.Blocks[len(bc.Blocks)-1]

	// Coinbase first
	rewardTx := NewCoinbaseTransaction(minerAddr, MiningReward)

	// Pull mempool txs
	txs := bc.Mempool.Flush()
	txs = append([]Transaction{rewardTx}, txs...)

	newBlock := Block{
		Index:        last.Index + 1,
		Timestamp:    now(),
		Transactions: txs,
		PrevHash:     last.Hash,
		Nonce:        0,
	}

	MineBlock(&newBlock)

	if !IsBlockValid(newBlock, last) {
		return Block{}, ErrInvalidBlock
	}

	// Apply to state
	if err := bc.State.ApplyBlock(newBlock); err != nil {
		return Block{}, err
	}

	bc.Blocks = append(bc.Blocks, newBlock)
	return newBlock, nil
}

func (bc *Blockchain) Balance(addr string) int {
	return bc.State.BalanceOf(addr)
}

// TryAddBlock is used by p2p: attempt to append a received block.
func (bc *Blockchain) TryAddBlock(b Block) bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if len(bc.Blocks) == 0 {
		return false
	}

	last := bc.Blocks[len(bc.Blocks)-1]
	if !IsBlockValid(b, last) {
		return false
	}

	// Apply state
	if err := bc.State.ApplyBlock(b); err != nil {
		return false
	}

	bc.Blocks = append(bc.Blocks, b)
	return true
}

// TryReplaceChain replaces local chain if the new one is valid and longer.
func (bc *Blockchain) TryReplaceChain(newChain []Block) bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if len(newChain) <= len(bc.Blocks) {
		return false
	}
	if !IsChainValid(newChain) {
		return false
	}

	// Rebuild state from scratch
	newState := NewState()
	for _, b := range newChain {
		if err := newState.ApplyBlock(b); err != nil {
			return false
		}
	}

	bc.Blocks = newChain
	bc.State = newState
	return true
}