package blockchain

import "time"

type Blockchain struct {
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
	// Apply genesis to state (usually empty, but keeps pipeline consistent)
	_ = bc.RebuildState()
	return bc
}

func (bc *Blockchain) RebuildState() error {
	bc.State = NewState()
	for _, b := range bc.Blocks {
		for _, tx := range b.Transactions {
			if ok, _ := tx.VerifySignatureOnly(); !ok {
				return ErrInvalidTransaction
			}
			if err := bc.State.ApplyTx(tx); err != nil {
				return err
			}
		}
	}
	return nil
}

func (bc *Blockchain) AddTransaction(tx Transaction) error {
	// signature validation
	ok, err := tx.VerifySignatureOnly()
	if err != nil || !ok {
		return ErrInvalidTransaction
	}

	// balance validation (prevent overspend)
	if !tx.IsCoinbase() {
		if bc.State.GetBalance(tx.From) < tx.Amount {
			return ErrInsufficientFunds
		}
	}

	return bc.Mempool.AddTransaction(tx)
}

// MinePendingTransactions now REQUIRES miner address to pay reward
func (bc *Blockchain) MinePendingTransactions(minerAddr string) (Block, error) {
	last := bc.Blocks[len(bc.Blocks)-1]

	// Add coinbase reward at top
	txs := []Transaction{NewCoinbaseTransaction(minerAddr)}
	txs = append(txs, bc.Mempool.Flush()...)

	newBlock := Block{
		Index:        last.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: txs,
		PrevHash:     last.Hash,
	}

	MineBlock(&newBlock)

	// Validate + apply to state
	if ok := bc.TryAddBlock(newBlock); !ok {
		return Block{}, ErrInvalidBlock
	}

	return newBlock, nil
}

func (bc *Blockchain) TryAddBlock(b Block) bool {
	last := bc.Blocks[len(bc.Blocks)-1]
	if !IsBlockValid(b, last) {
		return false
	}

	// Apply txs to state (including coinbase)
	// Work on a copy first
	tmp := NewState()
	for k, v := range bc.State.Balances {
		tmp.Balances[k] = v
	}

	for _, tx := range b.Transactions {
		ok, _ := tx.VerifySignatureOnly()
		if !ok {
			return false
		}
		if err := tmp.ApplyTx(tx); err != nil {
			return false
		}
	}

	bc.Blocks = append(bc.Blocks, b)
	bc.State = tmp
	return true
}

func (bc *Blockchain) TryReplaceChain(candidate []Block) bool {
	if len(candidate) <= len(bc.Blocks) {
		return false
	}
	if !IsChainValid(candidate) {
		return false
	}

	// Validate candidate against state rules by rebuilding
	oldBlocks := bc.Blocks
	bc.Blocks = candidate
	if err := bc.RebuildState(); err != nil {
		bc.Blocks = oldBlocks
		_ = bc.RebuildState()
		return false
	}
	return true
}
