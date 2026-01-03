package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"time"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/wallet"
)

type Transaction struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int    `json:"amount"`
	Timestamp int64  `json:"timestamp"`

	// Signing fields
	PublicKey []byte `json:"publicKey"`
	Signature []byte `json:"signature"`
}

func NewUnsignedTransaction(from, to string, amount int) Transaction {
	return Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
	}
}

// Hash returns a deterministic hash of the tx content (excluding signature).
func (tx Transaction) Hash() []byte {
	h := sha256.New()

	h.Write([]byte(tx.From))
	h.Write([]byte(tx.To))

	var amt [8]byte
	binary.BigEndian.PutUint64(amt[:], uint64(tx.Amount))
	h.Write(amt[:])

	var ts [8]byte
	binary.BigEndian.PutUint64(ts[:], uint64(tx.Timestamp))
	h.Write(ts[:])

	h.Write(tx.PublicKey)

	sum := h.Sum(nil)
	return sum
}

func (tx *Transaction) SignWith(w *wallet.Wallet) error {
	if w == nil {
		return errors.New("nil wallet")
	}

	// Ensure tx.From matches wallet address
	if tx.From != w.Address {
		return errors.New("tx.From does not match wallet address")
	}

	tx.PublicKey = w.PublicKey
	msgHash := tx.Hash()

	sig, err := w.Sign(msgHash)
	if err != nil {
		return err
	}

	tx.Signature = sig
	return nil
}

func (tx Transaction) Verify() (bool, error) {
	// Address must match public key
	derived := wallet.AddressFromPublicKey(tx.PublicKey)
	if derived != tx.From {
		return false, errors.New("from address does not match public key")
	}

	msgHash := tx.Hash()
	return wallet.VerifySignature(tx.PublicKey, msgHash, tx.Signature)
}
