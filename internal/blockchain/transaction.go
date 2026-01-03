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

	PublicKey []byte `json:"publicKey,omitempty"`
	Signature []byte `json:"signature,omitempty"`
}

func NewUnsignedTransaction(from, to string, amount int) Transaction {
	return Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
	}
}

func NewCoinbaseTransaction(to string) Transaction {
	return Transaction{
		From:      CoinbaseFrom,
		To:        to,
		Amount:    MinerReward,
		Timestamp: time.Now().Unix(),
		// no signature needed
	}
}

func (tx Transaction) IsCoinbase() bool {
	return tx.From == CoinbaseFrom
}

// Hash excludes Signature but includes PublicKey if present.
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

	return h.Sum(nil)
}

func (tx *Transaction) SignWith(w *wallet.Wallet) error {
	if tx.IsCoinbase() {
		return errors.New("coinbase tx cannot be signed")
	}
	if w == nil {
		return errors.New("nil wallet")
	}
	if tx.From != w.Address {
		return errors.New("tx.From does not match wallet address")
	}

	tx.PublicKey = w.PublicKey
	sig, err := w.Sign(tx.Hash())
	if err != nil {
		return err
	}
	tx.Signature = sig
	return nil
}

func (tx Transaction) VerifySignatureOnly() (bool, error) {
	if tx.IsCoinbase() {
		// coinbase has no signature
		return true, nil
	}

	if len(tx.PublicKey) == 0 || len(tx.Signature) == 0 {
		return false, errors.New("missing pubkey/signature")
	}

	derived := wallet.AddressFromPublicKey(tx.PublicKey)
	if derived != tx.From {
		return false, errors.New("from address does not match public key")
	}

	return wallet.VerifySignature(tx.PublicKey, tx.Hash(), tx.Signature)
}
