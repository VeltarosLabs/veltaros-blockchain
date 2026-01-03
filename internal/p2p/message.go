package p2p

import "encoding/json"

type MessageType string

const (
	MsgTransaction MessageType = "tx"
	MsgBlock       MessageType = "block"
	MsgGetChain    MessageType = "get_chain"
	MsgChain       MessageType = "chain"
	MsgMine        MessageType = "mine"
)

type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}
