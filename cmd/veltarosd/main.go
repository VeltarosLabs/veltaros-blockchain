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
	// If user passed only "4000", convert to ":4000"
	if !strings.Contains(s, ":") {
		return ":" + s
	}
	return s
}

func main() {
	// HTTP API
	addrFlag := flag.String("addr", "3000", "HTTP port to listen on (example: 3000 or :3000)")
	dataDir := flag.String("data", "data", "data directory (chain persistence)")

	// P2P
	p2pAddrFlag := flag.String("p2p", ":4000", "P2P listen address (example: :4000)")
	peerFlag := flag.String("peer", "", "Connect to peer (ip:port)")

	flag.Parse()

	// Normalize http port for network.Start(port)
	httpPort := strings.TrimPrefix(strings.TrimSpace(*addrFlag), ":")

	// Normalize p2p listen address
	p2pAddr := ensurePortHasColon(*p2pAddrFlag)

	// Load chain (or create new)
	bc, err := blockchain.LoadFromDisk(*dataDir)
	if err != nil {
		log.Fatal(err)
	}
	if bc == nil {
		bc = blockchain.NewBlockchain()
		_ = bc.SaveToDisk(*dataDir)
	}

	// P2P node
	p2pNode := p2p.NewNode(p2pAddr, bc)

	// Start P2P server in background
	go func() {
		if err := p2pNode.StartServer(); err != nil {
			log.Println("p2p server error:", err)
		}
	}()

	// Optionally connect to a peer
	if strings.TrimSpace(*peerFlag) != "" {
		if err := p2pNode.Connect(strings.TrimSpace(*peerFlag)); err != nil {
			log.Println("p2p connect error:", err)
		}
	}

	// HTTP API node
	api := network.NewNode(bc)
	api.DataDir = *dataDir

	// Wire broadcaster (HTTP actions -> P2P gossip)
	api.Broadcaster = p2pNode

	// Start HTTP API (blocks forever)
	api.Start(httpPort)
}
