package blockchain

// GenesisBlock creates the first block of the chain
func GenesisBlock() Block {
	block := Block{
		Index:        0,
		Timestamp:    0,
		Transactions: []Transaction{},
		PrevHash:     "0",
	}
	MineBlock(&block)
	return block
}
