package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
	"github.com/VeltarosLabs/veltaros-blockchain/internal/network"
)

func main() {
	addr := flag.String("addr", ":3000", "listen address (e.g. :3000)")
	flag.Parse()

	chain := blockchain.NewBlockchain()
	node := network.NewNode(chain)

	log.Println("Starting node on", *addr)
	node.Start(*addr)

	// Keep process alive (and cleanly fix the Read() mismatch)
	fmt.Println("Node running. Press ENTER to stop.")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}
