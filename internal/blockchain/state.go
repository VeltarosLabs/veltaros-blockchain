package blockchain

import "errors"

// State holds account balances
type State struct {
	Balances map[string]int
}

func NewState() *State {
	return &State{
		Balances: make(map[string]int),
	}
}

func (s *State) GetBalance(addr string) int {
	return s.Balances[addr]
}

func (s *State) Credit(addr string, amount int) {
	s.Balances[addr] += amount
}

func (s *State) Debit(addr string, amount int) error {
	if s.Balances[addr] < amount {
		return errors.New("insufficient funds")
	}
	s.Balances[addr] -= amount
	return nil
}

// ApplyTx applies a tx to balances (assumes signature already verified)
func (s *State) ApplyTx(tx Transaction) error {
	if tx.Amount <= 0 {
		return errors.New("amount must be > 0")
	}

	if tx.IsCoinbase() {
		s.Credit(tx.To, tx.Amount)
		return nil
	}

	// normal transfer
	if err := s.Debit(tx.From, tx.Amount); err != nil {
		return err
	}
	s.Credit(tx.To, tx.Amount)
	return nil
}
