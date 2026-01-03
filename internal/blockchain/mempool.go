package blockchain

import "sync"

type Mempool struct {
	mu  sync.Mutex
	txs []Transaction
}

func NewMempool() *Mempool {
	return &Mempool{txs: make([]Transaction, 0)}
}

func (m *Mempool) AddTransaction(tx Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txs = append(m.txs, tx)
}

// Flush returns all pending txs and clears the pool.
func (m *Mempool) Flush() []Transaction {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]Transaction, len(m.txs))
	copy(out, m.txs)
	m.txs = m.txs[:0]
	return out
}
