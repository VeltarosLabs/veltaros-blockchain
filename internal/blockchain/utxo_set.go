package blockchain

import (
	"fmt"
)

// UTXOSet is an in-memory view of unspent outputs.
// Later we can persist it for performance.
type UTXOSet struct {
	UTXOs map[OutPoint]TxOut `json:"utxos"`
}

func NewUTXOSet() *UTXOSet {
	return &UTXOSet{UTXOs: make(map[OutPoint]TxOut)}
}

// Rebuild scans the entire chain and reconstructs the UTXO set.
// This is Bitcoin-like (node can always rebuild from blocks).
func (u *UTXOSet) Rebuild(blocks []Block) {
	u.UTXOs = make(map[OutPoint]TxOut)

	// Track spent outputs
	spent := make(map[OutPoint]bool)

	// Walk blocks in order; for correctness, you can also walk from genesis to tip.
	for _, b := range blocks {
		// If your Block currently doesn't have UTXO txs yet, this will be empty until Step 2.
		for _, tx := range b.UTXOTxs {
			// mark inputs spent
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					spent[in.PrevOut] = true
				}
			}
			// add outputs
			for idx, out := range tx.Vout {
				op := OutPoint{TxID: tx.ID, Vout: idx}
				// only keep if not spent later in scan
				if !spent[op] {
					u.UTXOs[op] = out
				}
			}
		}
	}

	// Remove any outputs that were spent by later txs in scan
	for op := range spent {
		delete(u.UTXOs, op)
	}
}

// Balance sums all UTXOs locked to pubKeyHash.
func (u *UTXOSet) Balance(pubKeyHash []byte) int {
	sum := 0
	for _, out := range u.UTXOs {
		if samePubKeyHash(out.PubKeyHash, pubKeyHash) {
			sum += out.Value
		}
	}
	return sum
}

// FindSpendable selects UTXOs to cover amount (simple greedy).
func (u *UTXOSet) FindSpendable(pubKeyHash []byte, amount int) (int, map[OutPoint]TxOut) {
	acc := 0
	chosen := make(map[OutPoint]TxOut)

	for op, out := range u.UTXOs {
		if !samePubKeyHash(out.PubKeyHash, pubKeyHash) {
			continue
		}
		chosen[op] = out
		acc += out.Value
		if acc >= amount {
			break
		}
	}
	return acc, chosen
}

// ValidateTx checks that inputs exist in UTXO set and sum(inputs) >= sum(outputs).
// Signature/script validation comes next step.
func (u *UTXOSet) ValidateTx(tx UTXOTransaction) error {
	if tx.IsCoinbase() {
		if len(tx.Vout) == 0 {
			return fmt.Errorf("coinbase must have outputs")
		}
		return nil
	}

	inSum := 0
	seen := make(map[OutPoint]bool)

	for _, in := range tx.Vin {
		if in.PrevOut.TxID == "" || in.PrevOut.Vout < 0 {
			return fmt.Errorf("invalid input outpoint")
		}
		if seen[in.PrevOut] {
			return fmt.Errorf("double spend inside tx: %s", in.PrevOut.String())
		}
		seen[in.PrevOut] = true

		prev, ok := u.UTXOs[in.PrevOut]
		if !ok {
			return fmt.Errorf("input missing/unspent not found: %s", in.PrevOut.String())
		}
		inSum += prev.Value
	}

	outSum := 0
	for _, out := range tx.Vout {
		if out.Value <= 0 {
			return fmt.Errorf("invalid output value")
		}
		outSum += out.Value
	}

	if inSum < outSum {
		return fmt.Errorf("insufficient input sum: in=%d out=%d", inSum, outSum)
	}

	return nil
}

// ApplyTx updates the UTXO set (spend inputs, add outputs).
func (u *UTXOSet) ApplyTx(tx UTXOTransaction) error {
	if err := u.ValidateTx(tx); err != nil {
		return err
	}

	if !tx.IsCoinbase() {
		for _, in := range tx.Vin {
			delete(u.UTXOs, in.PrevOut)
		}
	}
	for idx, out := range tx.Vout {
		u.UTXOs[OutPoint{TxID: tx.ID, Vout: idx}] = out
	}
	return nil
}
