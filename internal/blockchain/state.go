package blockchain

import "sync"

type State struct {
	mu       sync.RWMutex
	balances map[string]int
}

func NewState() *State {
	return &State{balances: make(map[string]int)}
}

func (s *State) GetBalance(addr string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.balances[addr]
}

// BalanceOf is an alias used by some parts of the codebase.
func (s *State) BalanceOf(addr string) int {
	return s.GetBalance(addr)
}

// ApplyBlock applies all txs in the block to the state.
// Rule (simple):
// - if tx.From != "" then subtract Amount from From
// - always add Amount to tx.To
func (s *State) ApplyBlock(b Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, tx := range b.Transactions {
		if tx.Amount < 0 {
			// ignore/invalid; keep it simple
			continue
		}

		if tx.From != "" {
			s.balances[tx.From] -= tx.Amount
		}
		if tx.To != "" {
			s.balances[tx.To] += tx.Amount
		}
	}
	return nil
}
