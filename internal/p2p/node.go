package p2p

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
)

// Node represents a P2P node that can connect to peers and sync blocks/txs.
type Node struct {
	Address    string
	Blockchain *blockchain.Blockchain

	Peers map[string]*Peer
	lock  sync.Mutex
}

// NewNode creates a new P2P node.
func NewNode(address string, bc *blockchain.Blockchain) *Node {
	return &Node{
		Address:    address,
		Blockchain: bc,
		Peers:      make(map[string]*Peer),
	}
}

// Connect connects to a remote peer and starts listening to messages from it.
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

	// Ask peer for chain after connecting (sync)
	n.sendToPeer(peer, Message{Type: MsgGetChain, Data: json.RawMessage(`{}`)})

	fmt.Println("Connected to peer:", addr)
	return nil
}

// Broadcast sends msg to all connected peers.
func (n *Node) Broadcast(msg Message) {
	n.lock.Lock()
	defer n.lock.Unlock()

	for _, peer := range n.Peers {
		_ = json.NewEncoder(peer.Conn).Encode(msg)
	}
}

// BroadcastExcept sends msg to all peers except senderAddr.
func (n *Node) BroadcastExcept(senderAddr string, msg Message) {
	n.lock.Lock()
	defer n.lock.Unlock()

	for addr, peer := range n.Peers {
		if addr == senderAddr {
			continue
		}
		_ = json.NewEncoder(peer.Conn).Encode(msg)
	}
}

// sendToPeer sends msg to a specific peer.
func (n *Node) sendToPeer(peer *Peer, msg Message) {
	_ = json.NewEncoder(peer.Conn).Encode(msg)
}

// handlePeer reads messages in a loop until the peer disconnects.
func (n *Node) handlePeer(peer *Peer) {
	decoder := json.NewDecoder(peer.Conn)

	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			fmt.Println("Peer disconnected:", peer.Addr)

			n.lock.Lock()
			delete(n.Peers, peer.Addr)
			n.lock.Unlock()

			_ = peer.Conn.Close()
			return
		}

		n.handleMessage(peer, msg)
	}
}

// handleMessage processes an incoming message from a peer.
func (n *Node) handleMessage(peer *Peer, msg Message) {
	switch msg.Type {

	// Receive a transaction and add to mempool.
	case MsgTransaction:
		var tx blockchain.Transaction
		_ = json.Unmarshal(msg.Data, &tx)

		if err := n.Blockchain.AddTransaction(tx); err != nil {
			fmt.Println("Rejected tx:", err)
			return
		}

		// rebroadcast to others
		n.BroadcastExcept(peer.Addr, msg)

	// Mine request: mine current mempool, broadcast new block.
	case MsgMine:
		var payload struct {
			Miner string `json:"miner"`
		}
		_ = json.Unmarshal(msg.Data, &payload)

		if payload.Miner == "" {
			fmt.Println("Mine request missing miner address")
			return
		}

		newBlock, err := n.Blockchain.MinePendingTransactions(payload.Miner)
		if err != nil {
			fmt.Println("Mine failed:", err)
			return
		}

		raw, _ := json.Marshal(newBlock)
		n.Broadcast(Message{Type: MsgBlock, Data: raw})

	// Receive a block and try to append.
	case MsgBlock:
		var b blockchain.Block
		_ = json.Unmarshal(msg.Data, &b)

		// Try append; if fails, request full chain (fork/missing blocks)
		if ok := n.Blockchain.TryAddBlock(b); !ok {
			n.BroadcastExcept(peer.Addr, Message{Type: MsgGetChain, Data: json.RawMessage(`{}`)})
			return
		}

		// If accepted, rebroadcast
		n.BroadcastExcept(peer.Addr, msg)

	// Peer asks for our chain.
	case MsgGetChain:
		raw, _ := json.Marshal(n.Blockchain.Blocks)
		n.sendToPeer(peer, Message{Type: MsgChain, Data: raw})

	// Peer sends their chain, we prefer longer valid chain.
	case MsgChain:
		var chain []blockchain.Block
		_ = json.Unmarshal(msg.Data, &chain)

		if n.Blockchain.TryReplaceChain(chain) {
			fmt.Println("Chain updated from peer:", peer.Addr)
		}

	default:
		// Unknown message type: ignore
	}
}

// Broadcaster interface methods (used by internal/network HTTP API)
func (n *Node) BroadcastTx(tx blockchain.Transaction) {
	raw, err := json.Marshal(tx)
	if err != nil {
		return
	}
	n.Broadcast(Message{Type: MsgTransaction, Data: raw})
}

func (n *Node) BroadcastBlock(b blockchain.Block) {
	raw, err := json.Marshal(b)
	if err != nil {
		return
	}
	n.Broadcast(Message{Type: MsgBlock, Data: raw})
}
