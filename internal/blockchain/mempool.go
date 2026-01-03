package blockchain

import "sync"

// Mempool stores pending transactions
type Mempool struct {
	transactions []Transaction
	lock         sync.Mutex
}

// NewMempool creates a new mempool
func NewMempool() *Mempool {
	return &Mempool{
		transactions: []Transaction{},
	}
}

// AddTransaction adds tx to mempool
func (m *Mempool) AddTransaction(tx Transaction) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.transactions = append(m.transactions, tx)
}

// Flush returns all txs and clears mempool
func (m *Mempool) Flush() []Transaction {
	m.lock.Lock()
	defer m.lock.Unlock()

	txs := m.transactions
	m.transactions = []Transaction{}
	return txs
}
