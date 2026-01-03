package blockchain

// IsBlockValid validates a block against previous block + PoW + tx signatures.
func IsBlockValid(newBlock Block, prevBlock Block) bool {
	if prevBlock.Index+1 != newBlock.Index {
		return false
	}
	if prevBlock.Hash != newBlock.PrevHash {
		return false
	}

	calculated := CalculateBlockHash(&newBlock)
	if newBlock.Hash != calculated {
		return false
	}

	// PoW check (IsPoWValid is defined in mining.go)
	if !IsPoWValid(newBlock.Hash) {
		return false
	}

	// Validate transactions (signature-only; coinbase allowed)
	for _, tx := range newBlock.Transactions {
		ok, _ := tx.VerifySignatureOnly()
		if !ok {
			return false
		}
	}

	return true
}

// IsChainValid validates the entire chain.
func IsChainValid(chain []Block) bool {
	if len(chain) == 0 {
		return false
	}

	for i := 1; i < len(chain); i++ {
		if !IsBlockValid(chain[i], chain[i-1]) {
			return false
		}
	}

	return true
}
