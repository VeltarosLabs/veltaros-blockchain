package p2p

import (
	"fmt"
	"net"
)

// StartServer starts listening for peers
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

		peer := &Peer{
			Conn: conn,
			Addr: conn.RemoteAddr().String(),
		}

		n.lock.Lock()
		n.Peers[peer.Addr] = peer
		n.lock.Unlock()

		go n.handlePeer(peer)
	}
}
