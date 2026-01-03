package wallet

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

func SavePrivateKeyPEM(path string, priv *ecdsa.PrivateKey) error {
	if priv == nil {
		return errors.New("nil private key")
	}

	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return err
	}

	block := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	}

	return os.WriteFile(path, pem.EncodeToMemory(block), 0600)
}

func LoadPrivateKeyPEM(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, errors.New("invalid PEM file")
	}

	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return priv, nil
}
