package p2p

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
)

// Node represents a blockchain network node
type Node struct {
	Address    string
	Blockchain *blockchain.Blockchain

	Peers map[string]*Peer
	lock  sync.Mutex
}

// NewNode creates a new P2P node
func NewNode(address string, bc *blockchain.Blockchain) *Node {
	return &Node{
		Address:    address,
		Blockchain: bc,
		Peers:      make(map[string]*Peer),
	}
}

// Connect connects to another peer
func (n *Node) Connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	peer := &Peer{Conn: conn, Addr: addr}

	n.lock.Lock()
	n.Peers[addr] = peer
	n.lock.Unlock()

	go n.handlePeer(peer)
	fmt.Println("Connected to peer:", addr)
	return nil
}

// Broadcast sends message to all peers
func (n *Node) Broadcast(msg Message) {
	n.lock.Lock()
	defer n.lock.Unlock()

	for _, peer := range n.Peers {
		_ = json.NewEncoder(peer.Conn).Encode(msg)
	}
}

// handlePeer listens for peer messages
func (n *Node) handlePeer(peer *Peer) {
	decoder := json.NewDecoder(peer.Conn)

	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			fmt.Println("Peer disconnected:", peer.Addr)
			return
		}
		n.handleMessage(msg)
	}
}

// handleMessage processes incoming messages
func (n *Node) handleMessage(msg Message) {
	switch msg.Type {

	case MsgTransaction:
		var tx blockchain.Transaction
		_ = json.Unmarshal(msg.Data, &tx)
		n.Blockchain.AddTransaction(tx)

	case MsgBlock:
		var block blockchain.Block
		_ = json.Unmarshal(msg.Data, &block)
		n.Blockchain.Blocks = append(n.Blockchain.Blocks, block)
	}
}
