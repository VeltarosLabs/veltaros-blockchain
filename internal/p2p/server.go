package p2p

import (
	"fmt"
	"net"
)

func (n *Node) StartServer() error {
	ln, err := net.Listen("tcp", n.Address)
	if err != nil {
		return err
	}

	fmt.Println("Node listening on", n.Address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		remote := conn.RemoteAddr().String()
		peer := &Peer{Conn: conn, Addr: remote}

		n.lock.Lock()
		n.Peers[remote] = peer
		n.lock.Unlock()

		go n.handlePeer(peer)

		// Ask the inbound peer for their chain as well
		n.sendToPeer(peer, Message{Type: MsgGetChain, Data: []byte(`{}`)})
	}
}
