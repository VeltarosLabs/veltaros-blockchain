package network

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
)

type Node struct {
	Chain       *blockchain.Blockchain
	Broadcaster Broadcaster
	DataDir     string
}

func NewNode(chain *blockchain.Blockchain) *Node {
	return &Node{Chain: chain}
}

// wrap adds panic recovery + consistent JSON headers.
func (n *Node) wrap(h func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC in %s %s: %v\n%s", r.Method, r.URL.Path, rec, string(debug.Stack()))
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		h(w, r)
	}
}

func (n *Node) Start(port string) {
	mux := http.NewServeMux()

	// Keep old routes + new routes (so your CLI keeps working)
	mux.HandleFunc("/transaction", n.wrap(n.handleTransaction)) // POST (full tx json)
	mux.HandleFunc("/tx", n.wrap(n.handleNewTx))                // POST (from,to,amount)
	mux.HandleFunc("/mine", n.wrap(n.handleMine))               // POST (miner)
	mux.HandleFunc("/chain", n.wrap(n.handleChain))             // GET
	mux.HandleFunc("/balance", n.wrap(n.handleBalance))         // GET ?addr=
	mux.HandleFunc("/nonce", n.wrap(n.handleNonce))             // GET ?addr=

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("HTTP API listening on port", port)
	log.Fatal(srv.ListenAndServe())
}

// POST /transaction
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

	if n.DataDir != "" {
		_ = n.Chain.SaveToDisk(n.DataDir)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

// POST /tx
func (n *Node) handleNewTx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount int    `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if req.From == "" || req.To == "" || req.Amount <= 0 {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	tx := blockchain.NewTransaction(req.From, req.To, req.Amount, 0, 0)
	if err := n.Chain.AddTransaction(tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if n.Broadcaster != nil {
		n.Broadcaster.BroadcastTx(tx)
	}

	if n.DataDir != "" {
		_ = n.Chain.SaveToDisk(n.DataDir)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok": true,
		"tx": tx,
	})
}

// POST /mine
// body: {"miner":"ADDRESS"}
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

	// Mine first, then persist (avoid saving broken state if mining fails)
	block, err := n.Chain.MinePendingTransactions(payload.Miner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if n.Broadcaster != nil {
		n.Broadcaster.BroadcastBlock(block)
	}

	if n.DataDir != "" {
		_ = n.Chain.SaveToDisk(n.DataDir)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(block)
}

// GET /chain
func (n *Node) handleChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(n.Chain.Blocks)
}

// GET /balance?addr=ADDRESS
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

func (n *Node) handleNonce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	addr := r.URL.Query().Get("addr")
	if addr == "" {
		http.Error(w, "missing addr", http.StatusBadRequest)
		return
	}

	nonce := n.Chain.State.NextNonce(addr)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"address": addr,
		"nonce":   nonce,
	})
}
