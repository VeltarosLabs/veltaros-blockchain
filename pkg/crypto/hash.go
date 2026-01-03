package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

func GenerateHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
