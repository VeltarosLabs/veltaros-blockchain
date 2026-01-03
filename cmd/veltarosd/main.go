package main

import (
	"fmt"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
)

func main() {
	fmt.Println("Veltaros Blockchain Node Starting...")

	chain := blockchain.NewBlockchain()
	chain.AddBlock("First transaction")
	chain.AddBlock("Second transaction")

	for _, block := range chain.Blocks {
		fmt.Printf("\nBlock %d\n", block.Index)
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("PrevHash: %s\n", block.PrevHash)
		fmt.Printf("Hash: %s\n", block.Hash)
	}
}
