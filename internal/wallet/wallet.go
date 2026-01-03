package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
	Address    string
}

func New() (*Wallet, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	pub := marshalPublicKey(&priv.PublicKey)
	addr := AddressFromPublicKey(pub)

	return &Wallet{
		PrivateKey: priv,
		PublicKey:  pub,
		Address:    addr,
	}, nil
}

func AddressFromPublicKey(pub []byte) string {
	// Simple address: first 20 bytes of SHA256(pub) (40 hex chars)
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:20])
}

func (w *Wallet) Sign(messageHash []byte) ([]byte, error) {
	if w == nil || w.PrivateKey == nil {
		return nil, errors.New("wallet has no private key")
	}
	// ASN.1 DER signature
	sig, err := ecdsa.SignASN1(rand.Reader, w.PrivateKey, messageHash)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func VerifySignature(pubKeyBytes []byte, messageHash []byte, sig []byte) (bool, error) {
	pub, err := unmarshalPublicKey(pubKeyBytes)
	if err != nil {
		return false, err
	}
	ok := ecdsa.VerifyASN1(pub, messageHash, sig)
	return ok, nil
}

func marshalPublicKey(pub *ecdsa.PublicKey) []byte {
	// Uncompressed form: 0x04 || X || Y
	return elliptic.Marshal(elliptic.P256(), pub.X, pub.Y)
}

func unmarshalPublicKey(b []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(elliptic.P256(), b)
	if x == nil || y == nil {
		return nil, errors.New("invalid public key bytes")
	}
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
}

// NewWallet is kept for CLI compatibility.
func NewWallet() (*Wallet, error) {
	return New() // <-- change this to whatever your constructor is called
}
