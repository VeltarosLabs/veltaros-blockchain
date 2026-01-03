package blockchain

// IsBlockValid checks linkage + PoW + hash correctness.
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
	if !IsPoWValid(newBlock.Hash) {
		return false
	}

	// Optional: verify all tx signatures (recommended)
	for _, tx := range newBlock.Transactions {
		ok, _ := tx.Verify()
		if !ok {
			return false
		}
	}

	return true
}

// IsChainValid validates a full chain.
func IsChainValid(chain []Block) bool {
	if len(chain) == 0 {
		return false
	}
	// Genesis: basic PoW/hash validity
	gen := chain[0]
	if gen.PrevHash != "0" {
		return false
	}
	if gen.Hash != CalculateBlockHash(&gen) || !IsPoWValid(gen.Hash) {
		return false
	}

	for i := 1; i < len(chain); i++ {
		if !IsBlockValid(chain[i], chain[i-1]) {
			return false
		}
	}
	return true
}
