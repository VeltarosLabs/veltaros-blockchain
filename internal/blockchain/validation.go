package blockchain

func IsBlockValid(newBlock, previousBlock Block) bool {
	if previousBlock.Index+1 != newBlock.Index {
		return false
	}

	if previousBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateBlockHash(&newBlock) != newBlock.Hash {
		return false
	}

	return true
}
