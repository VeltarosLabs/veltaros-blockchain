package p2p

import "encoding/json"

// MessageType defines network message kinds
type MessageType string

const (
	MsgTransaction MessageType = "tx"
	MsgBlock       MessageType = "block"
)

// Message is exchanged between peers
type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}
