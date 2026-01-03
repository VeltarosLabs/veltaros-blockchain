package blockchain

import (
	"errors"
	"sync"
)

type Mempool struct {
	transactions []Transaction
	lock         sync.Mutex
}

func NewMempool() *Mempool {
	return &Mempool{transactions: []Transaction{}}
}

func (m *Mempool) AddTransaction(tx Transaction) error {
	ok, err := tx.Verify()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("invalid signature")
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.transactions = append(m.transactions, tx)
	return nil
}

func (m *Mempool) Flush() []Transaction {
	m.lock.Lock()
	defer m.lock.Unlock()

	txs := m.transactions
	m.transactions = []Transaction{}
	return txs
}
