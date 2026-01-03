package blockchain

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

	return true
}

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
