package p2p

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
)

type Node struct {
	Address    string
	Blockchain *blockchain.Blockchain

	Peers map[string]*Peer
	lock  sync.Mutex
}

func NewNode(address string, bc *blockchain.Blockchain) *Node {
	return &Node{
		Address:    address,
		Blockchain: bc,
		Peers:      make(map[string]*Peer),
	}
}

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

	// Ask for chain after connecting (sync)
	n.sendToPeer(peer, Message{Type: MsgGetChain, Data: json.RawMessage(`{}`)})

	fmt.Println("Connected to peer:", addr)
	return nil
}

func (n *Node) Broadcast(msg Message) {
	n.lock.Lock()
	defer n.lock.Unlock()

	for _, peer := range n.Peers {
		_ = json.NewEncoder(peer.Conn).Encode(msg)
	}
}

func (n *Node) BroadcastExcept(sender string, msg Message) {
	n.lock.Lock()
	defer n.lock.Unlock()

	for addr, peer := range n.Peers {
		if addr == sender {
			continue
		}
		_ = json.NewEncoder(peer.Conn).Encode(msg)
	}
}

func (n *Node) sendToPeer(peer *Peer, msg Message) {
	_ = json.NewEncoder(peer.Conn).Encode(msg)
}

func (n *Node) handlePeer(peer *Peer) {
	decoder := json.NewDecoder(peer.Conn)

	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			fmt.Println("Peer disconnected:", peer.Addr)
			n.lock.Lock()
			delete(n.Peers, peer.Addr)
			n.lock.Unlock()
			return
		}
		n.handleMessage(peer.Addr, msg)
	}
}

func (n *Node) handleMessage(sender string, msg Message) {
	switch msg.Type {

	case MsgTransaction:
		var tx blockchain.Transaction
		_ = json.Unmarshal(msg.Data, &tx)

		if err := n.Blockchain.AddTransaction(tx); err != nil {
			fmt.Println("Rejected tx:", err)
			return
		}

		// rebroadcast to others
		n.BroadcastExcept(sender, msg)

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

	case MsgBlock:
		var b blockchain.Block
		_ = json.Unmarshal(msg.Data, &b)

		// Try append; if fails, request chain (fork or missing blocks)
		if ok := n.Blockchain.TryAddBlock(b); !ok {
			n.BroadcastExcept(sender, Message{Type: MsgGetChain, Data: json.RawMessage(`{}`)})
			return
		}

		// If accepted, rebroadcast
		n.BroadcastExcept(sender, msg)

	case MsgGetChain:
		raw, _ := json.Marshal(n.Blockchain.Blocks)
		n.sendToPeer(n.Peers[sender], Message{Type: MsgChain, Data: raw})

	case MsgChain:
		var chain []blockchain.Block
		_ = json.Unmarshal(msg.Data, &chain)

		// Fork rule: prefer longer valid chain
		if n.Blockchain.TryReplaceChain(chain) {
			fmt.Println("Chain updated from peer:", sender)
			// After switching, you can optionally rebroadcast the new chain (not required)
		}
	}
}
