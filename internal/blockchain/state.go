package blockchain

import "errors"

type State struct {
	Balances map[string]int    `json:"balances"`
	Nonces   map[string]uint64 `json:"nonces"`
}

func NewState() *State {
	return &State{
		Balances: make(map[string]int),
		Nonces:   make(map[string]uint64),
	}
}

// --- Canonical methods ---

func (s *State) Balance(addr string) int {
	return s.Balances[addr]
}

func (s *State) NextNonce(addr string) uint64 {
	return s.Nonces[addr] + 1
}

// ApplyTransaction mutates state for a single tx.
// Must be called in deterministic order (block order).
func (s *State) ApplyTransaction(tx Transaction) error {
	if tx.Amount <= 0 {
		return errors.New("amount must be > 0")
	}
	if tx.To == "" {
		return errors.New("missing recipient")
	}

	// Coinbase (mining reward): From == ""
	if tx.IsCoinbase() {
		s.Balances[tx.To] += tx.Amount
		return nil
	}

	// Normal tx: verify signature
	if err := tx.Verify(); err != nil {
		return err
	}

	fee := tx.Fee
	if fee < 0 {
		return errors.New("fee must be >= 0")
	}

	required := tx.Amount + fee
	if s.Balances[tx.From] < required {
		return errors.New("insufficient balance")
	}

	// Nonce check (replay protection)
	expected := s.NextNonce(tx.From)
	if tx.Nonce != expected {
		return errors.New("bad nonce")
	}

	// Apply
	s.Balances[tx.From] -= required
	s.Balances[tx.To] += tx.Amount
	s.Nonces[tx.From] = tx.Nonce

	return nil
}

// ApplyBlock applies all transactions in the block in order.
// This exists because your blockchain.go currently calls bc.State.ApplyBlock(...).
func (s *State) ApplyBlock(b Block) error {
	for _, tx := range b.Transactions {
		if err := s.ApplyTransaction(tx); err != nil {
			return err
		}
	}
	return nil
}

// --- Compatibility methods (fix your compile errors without touching blockchain.go) ---

func (s *State) GetBalance(addr string) int { // used by internal/network/node.go in your screenshot
	return s.Balance(addr)
}

func (s *State) BalanceOf(addr string) int { // used by internal/blockchain/blockchain.go in your screenshot
	return s.Balance(addr)
}
