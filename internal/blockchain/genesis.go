package blockchain

func GenesisBlock() Block {
	gen := Block{
		Index:        0,
		Timestamp:    now(),
		Transactions: []Transaction{},
		PrevHash:     "",
		Nonce:        0,
	}
	MineBlock(&gen)
	return gen
}
