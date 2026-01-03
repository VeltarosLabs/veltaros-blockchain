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
	http.HandleFunc("/transaction", n.handleTransaction) // POST
	http.HandleFunc("/mine", n.handleMine)               // POST
	http.HandleFunc("/chain", n.handleChain)             // GET
	http.HandleFunc("/balance", n.handleBalance)         // GET ?addr=

	log.Println("HTTP API listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func (n *Node) handleTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var tx blockchain.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if err := n.Chain.AddTransaction(tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok": true,
	})
}

func (n *Node) handleMine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Miner string `json:"miner"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if payload.Miner == "" {
		http.Error(w, "missing miner address", http.StatusBadRequest)
		return
	}

	block, err := n.Chain.MinePendingTransactions(payload.Miner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(block)
}

func (n *Node) handleChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(n.Chain.Blocks)
}

func (n *Node) handleBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	addr := r.URL.Query().Get("addr")
	if addr == "" {
		http.Error(w, "missing addr", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"address": addr,
		"balance": n.Chain.State.GetBalance(addr),
		"unit":    "VLT",
	})
}
