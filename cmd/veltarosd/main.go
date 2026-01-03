package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
	"github.com/VeltarosLabs/veltaros-blockchain/internal/network"
	"github.com/VeltarosLabs/veltaros-blockchain/internal/p2p"
)

func main() {
	p2pPort := flag.String("p2p", "3000", "P2P port")
	httpPort := flag.String("http", "8080", "HTTP API port")
	peer := flag.String("peer", "", "Optional peer address host:port")
	flag.Parse()

	bc := blockchain.NewBlockchain()

	// Start P2P node
	node := p2p.NewNode(":"+*p2pPort, bc)
	go func() {
		if err := node.StartServer(); err != nil {
			fmt.Println("P2P server error:", err)
		}
	}()

	time.Sleep(250 * time.Millisecond)

	if *peer != "" {
		if err := node.Connect(*peer); err != nil {
			fmt.Println("Failed to connect to peer:", err)
		}
	}

	// Start HTTP debug API
	api := network.NewNode(bc)
	go api.Start(*httpPort)

	fmt.Println("veltarosd running")
	fmt.Println("P2P  : localhost:" + *p2pPort)
	fmt.Println("HTTP : localhost:" + *httpPort)
	fmt.Println("(CTRL+C to stop)")

	select {}
}
