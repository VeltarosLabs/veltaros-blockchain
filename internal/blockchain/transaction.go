package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"time"
)

// Transaction represents an account-based transfer.
// Coinbase (mining reward) is represented by From == "" and Nonce == 0.
type Transaction struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int    `json:"amount"`
	Fee       int    `json:"fee,omitempty"`
	Nonce     uint64 `json:"nonce,omitempty"`
	Timestamp int64  `json:"timestamp"`

	// Sender authorization:
	// PubKey is uncompressed format 04||X||Y (65 bytes) encoded hex.
	// Sig is ASN.1 DER signature encoded hex.
	PubKey string `json:"pubKey,omitempty"`
	Sig    string `json:"sig,omitempty"`
}

func NewTransaction(from, to string, amount, fee int, nonce uint64) Transaction {
	tx := Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Fee:       fee,
		Nonce:     nonce,
		Timestamp: time.Now().Unix(),
	}
	tx.ID = tx.computeID()
	return tx
}

// For coinbase (mining reward)
func NewCoinbase(to string, amount int) Transaction {
	tx := Transaction{
		From:      "",
		To:        to,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
	}
	tx.ID = tx.computeID()
	return tx
}

// Compatibility: some of your code still calls NewCoinbaseTransaction(...)
func NewCoinbaseTransaction(to string, amount int) Transaction {
	return NewCoinbase(to, amount)
}

func (tx Transaction) IsCoinbase() bool {
	return tx.From == ""
}

func (tx Transaction) computeID() string {
	sum := sha256.Sum256(tx.signingBytes())
	return hex.EncodeToString(sum[:])
}

// signingBytes returns canonical bytes that are signed/verified.
// IMPORTANT: excludes Sig and ID to avoid circular hashing.
func (tx Transaction) signingBytes() []byte {
	type signable struct {
		From      string `json:"from"`
		To        string `json:"to"`
		Amount    int    `json:"amount"`
		Fee       int    `json:"fee,omitempty"`
		Nonce     uint64 `json:"nonce,omitempty"`
		Timestamp int64  `json:"timestamp"`
		PubKey    string `json:"pubKey,omitempty"`
	}
	b, _ := json.Marshal(signable{
		From:      tx.From,
		To:        tx.To,
		Amount:    tx.Amount,
		Fee:       tx.Fee,
		Nonce:     tx.Nonce,
		Timestamp: tx.Timestamp,
		PubKey:    tx.PubKey,
	})
	return b
}

// AddressFromPubKey derives the address from a public key.
// address = hex( sha256(pubKeyBytes)[:20] )
func AddressFromPubKey(pub ecdsa.PublicKey) string {
	pubBytes := MarshalPubKey(pub)
	sum := sha256.Sum256(pubBytes)
	return hex.EncodeToString(sum[:20])
}

func MarshalPubKey(pub ecdsa.PublicKey) []byte {
	xb := pub.X.Bytes()
	yb := pub.Y.Bytes()

	xp := make([]byte, 32)
	yp := make([]byte, 32)
	copy(xp[32-len(xb):], xb)
	copy(yp[32-len(yb):], yb)

	out := make([]byte, 0, 65)
	out = append(out, 0x04)
	out = append(out, xp...)
	out = append(out, yp...)
	return out
}

func UnmarshalPubKeyHex(pubHex string) (ecdsa.PublicKey, error) {
	raw, err := hex.DecodeString(pubHex)
	if err != nil {
		return ecdsa.PublicKey{}, err
	}
	if len(raw) != 65 || raw[0] != 0x04 {
		return ecdsa.PublicKey{}, errors.New("invalid pubkey format")
	}
	x := new(big.Int).SetBytes(raw[1:33])
	y := new(big.Int).SetBytes(raw[33:65])
	pub := ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
	if !pub.Curve.IsOnCurve(pub.X, pub.Y) {
		return ecdsa.PublicKey{}, errors.New("pubkey not on curve")
	}
	return pub, nil
}

func (tx Transaction) Verify() error {
	if tx.IsCoinbase() {
		return nil
	}

	if tx.PubKey == "" || tx.Sig == "" {
		return errors.New("missing pubKey or sig")
	}

	pub, err := UnmarshalPubKeyHex(tx.PubKey)
	if err != nil {
		return err
	}

	addr := AddressFromPubKey(pub)
	if !bytes.Equal([]byte(addr), []byte(tx.From)) {
		return errors.New("pubkey does not match from address")
	}

	sigBytes, err := hex.DecodeString(tx.Sig)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(tx.signingBytes())
	if !ecdsa.VerifyASN1(&pub, hash[:], sigBytes) {
		return errors.New("invalid signature")
	}

	if tx.ID != tx.computeID() {
		return errors.New("invalid tx id")
	}

	return nil
}

func (tx *Transaction) Sign(priv *ecdsa.PrivateKey) error {
	if tx.IsCoinbase() {
		return nil
	}

	tx.PubKey = hex.EncodeToString(MarshalPubKey(priv.PublicKey))
	tx.From = AddressFromPubKey(priv.PublicKey)

	tx.ID = tx.computeID()

	hash := sha256.Sum256(tx.signingBytes())
	sig, err := ecdsa.SignASN1(nil, priv, hash[:])
	if err != nil {
		return err
	}
	tx.Sig = hex.EncodeToString(sig)

	tx.ID = tx.computeID()
	return nil
}
