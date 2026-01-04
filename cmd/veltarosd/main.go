package main

import (
	"flag"
	"log"
	"strings"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
	"github.com/VeltarosLabs/veltaros-blockchain/internal/network"
)

func main() {
	addr := flag.String("addr", "3000", "port to listen on (example: 3000)")
	flag.Parse()

	port := *addr
	port = strings.TrimPrefix(port, ":")

	bc := blockchain.NewBlockchain()
	n := network.NewNode(bc)

	log.Println("Starting node on", port)
	n.Start(port)
}
