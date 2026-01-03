package p2p

import (
	"net"
)

// Peer represents a connected node
type Peer struct {
	Conn net.Conn
	Addr string
}
