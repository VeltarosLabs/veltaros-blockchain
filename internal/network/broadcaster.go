package network

import "github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"

// Broadcaster is implemented by the P2P node (or nil if P2P disabled).
type Broadcaster interface {
	BroadcastTx(tx blockchain.Transaction)
	BroadcastBlock(b blockchain.Block)
}
