package main

import (
	"flag"
	"log"
	"strings"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
	"github.com/VeltarosLabs/veltaros-blockchain/internal/network"
	"github.com/VeltarosLabs/veltaros-blockchain/internal/p2p"
)

func ensurePortHasColon(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if strings.HasPrefix(s, ":") {
		return s
	}
	// if user passed only "4000", convert to ":4000"
	if !strings.Contains(s, ":") {
		return ":" + s
	}
	return s
}

func main() {
	// HTTP API port (existing)
	addrFlag := flag.String("addr", "3000", "HTTP port to listen on (example: 3000 or :3000)")

	// P2P settings
	p2pAddrFlag := flag.String("p2p", ":4000", "P2P listen address (example: :4000)")
	peerFlag := flag.String("peer", "", "Connect to peer (ip:port)")

	flag.Parse()

	// Clean HTTP port: allow "3000" or ":3000"
	httpPort := strings.TrimPrefix(*addrFlag, ":")

	// Clean P2P listen address: allow "4000" or ":4000"
	p2pAddr := ensurePortHasColon(*p2pAddrFlag)

	// Blockchain
	bc := blockchain.NewBlockchain()

	// P2P node (tcp)
	p2pNode := p2p.NewNode(p2pAddr, bc)

	// Start P2P server in background
	go func() {
		if err := p2pNode.StartServer(); err != nil {
			log.Println("p2p server error:", err)
		}
	}()

	// Optional: connect to peer if provided
	if *peerFlag != "" {
		if err := p2pNode.Connect(*peerFlag); err != nil {
			log.Println("p2p connect error:", err)
		}
	}

	// HTTP API node
	api := network.NewNode(bc)

	// If your network Node has Broadcaster field, keep this line:
	api.Broadcaster = p2pNode

	// Start HTTP API (blocks forever)
	api.Start(httpPort)
}
