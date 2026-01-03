package main

import (
	"fmt"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
	"github.com/VeltarosLabs/veltaros-blockchain/internal/transaction"
)

func main() {
	fmt.Println("Veltaros Blockchain Node Starting...")

	bc := blockchain.NewBlockchain()

	tx1 := transaction.NewTransaction("Alice", "Bob", 50)
	tx2 := transaction.NewTransaction("Bob", "Charlie", 25)

	bc.AddBlock([]transaction.Transaction{tx1})
	bc.AddBlock([]transaction.Transaction{tx2})

	for _, block := range bc.Chain {
		fmt.Printf("\nBlock %d\n", block.Index)
		fmt.Printf("Hash: %s\n", block.Hash)
		fmt.Printf("PrevHash: %s\n", block.PrevHash)
		fmt.Printf("Transactions: %v\n", block.Transactions)
	}
}
