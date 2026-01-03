package network

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
)

type Node struct {
	Chain *blockchain.Blockchain
}

func NewNode(chain *blockchain.Blockchain) *Node {
	return &Node{Chain: chain}
}

func (n *Node) Start(port string) {
	http.HandleFunc("/transaction", n.handleTransaction)
	http.HandleFunc("/mine", n.handleMine)
	http.HandleFunc("/chain", n.handleChain)

	log.Println("Node listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func (n *Node) handleTransaction(w http.ResponseWriter, r *http.Request) {
	var tx blockchain.Transaction
	json.NewDecoder(r.Body).Decode(&tx)
	n.Chain.AddTransaction(tx)
	w.Write([]byte("Transaction added"))
}

func (n *Node) handleMine(w http.ResponseWriter, r *http.Request) {
	n.Chain.MinePendingTransactions()
	w.Write([]byte("Block mined"))
}

func (n *Node) handleChain(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(n.Chain.Blocks)
}
