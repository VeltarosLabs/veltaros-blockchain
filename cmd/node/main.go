package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
	"github.com/VeltarosLabs/veltaros-blockchain/internal/p2p"
	"github.com/VeltarosLabs/veltaros-blockchain/internal/wallet"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "start":
		startNode()
	case "wallet-new":
		createWallet()
	case "tx":
		sendSignedTransaction()
	case "mine":
		mineRemote()
	case "balance":
		balance()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Print(
		`Veltaros CLI (P2P)

Commands:
  start --port 3000 [--peer localhost:3001]             Start P2P node
  wallet-new --out wallet.pem                            Create wallet PEM
  tx --wallet wallet.pem --to ADDR --amount 10 --node localhost:3000
                                                         Send signed tx to node
  mine --wallet wallet.pem --node localhost:3000          Ask node to mine (reward to wallet)
  balance --addr ADDR --node localhost:3000               Get balance (rebuild from chain)
`)
}

func startNode() {
	cmd := flag.NewFlagSet("start", flag.ExitOnError)
	port := cmd.String("port", "3000", "Node port")
	peer := cmd.String("peer", "", "Peer address (optional)")
	cmd.Parse(os.Args[2:])

	bc := blockchain.NewBlockchain()
	node := p2p.NewNode(":"+*port, bc)

	go func() {
		if err := node.StartServer(); err != nil {
			fmt.Println("P2P server error:", err)
		}
	}()

	time.Sleep(250 * time.Millisecond)

	if *peer != "" {
		if err := node.Connect(*peer); err != nil {
			fmt.Println("Failed to connect:", err)
		}
	}

	fmt.Println("P2P node running on port", *port, "(CTRL+C to stop)")
	select {}
}

func createWallet() {
	cmd := flag.NewFlagSet("wallet-new", flag.ExitOnError)
	out := cmd.String("out", "wallet.pem", "Output wallet PEM file")
	cmd.Parse(os.Args[2:])

	w, err := wallet.New()
	if err != nil {
		fmt.Println("Wallet error:", err)
		return
	}

	if err := wallet.SavePrivateKeyPEM(*out, w.PrivateKey); err != nil {
		fmt.Println("Save error:", err)
		return
	}

	fmt.Println("Wallet created!")
	fmt.Println("Address:", w.Address)
	fmt.Println("Saved private key to:", *out)
}

func sendSignedTransaction() {
	cmd := flag.NewFlagSet("tx", flag.ExitOnError)
	walletPath := cmd.String("wallet", "wallet.pem", "Wallet PEM file")
	to := cmd.String("to", "", "Recipient address")
	amount := cmd.Int("amount", 0, "Amount")
	nodeAddr := cmd.String("node", "localhost:3000", "Node address host:port")
	cmd.Parse(os.Args[2:])

	if *to == "" || *amount <= 0 {
		fmt.Println("Usage: tx --wallet wallet.pem --to ADDR --amount 10 --node localhost:3000")
		return
	}

	priv, err := wallet.LoadPrivateKeyPEM(*walletPath)
	if err != nil {
		fmt.Println("Load wallet error:", err)
		return
	}

	w, err := walletFromPrivateKey(priv)
	if err != nil {
		fmt.Println("Wallet rebuild error:", err)
		return
	}

	tx := blockchain.NewUnsignedTransaction(w.Address, *to, *amount)
	if err := tx.SignWith(w); err != nil {
		fmt.Println("Sign error:", err)
		return
	}

	if err := sendTxToNode(*nodeAddr, tx); err != nil {
		fmt.Println("Send error:", err)
		return
	}

	fmt.Println("Signed transaction sent to node:", *nodeAddr)
	fmt.Println("From:", tx.From)
	fmt.Println("To:", tx.To)
	fmt.Println("Amount:", tx.Amount)
}

func mineRemote() {
	cmd := flag.NewFlagSet("mine", flag.ExitOnError)
	walletPath := cmd.String("wallet", "wallet.pem", "Wallet PEM file (miner reward)")
	nodeAddr := cmd.String("node", "localhost:3000", "Node address host:port")
	cmd.Parse(os.Args[2:])

	priv, err := wallet.LoadPrivateKeyPEM(*walletPath)
	if err != nil {
		fmt.Println("Load wallet error:", err)
		return
	}

	w, err := walletFromPrivateKey(priv)
	if err != nil {
		fmt.Println("Wallet rebuild error:", err)
		return
	}

	if err := sendMineToNode(*nodeAddr, w.Address); err != nil {
		fmt.Println("Mine request failed:", err)
		return
	}

	fmt.Println("Mine request sent to node:", *nodeAddr)
	fmt.Println("Miner reward address:", w.Address)
}

func balance() {
	cmd := flag.NewFlagSet("balance", flag.ExitOnError)
	addr := cmd.String("addr", "", "Address to check")
	nodeAddr := cmd.String("node", "localhost:3000", "Node address host:port")
	cmd.Parse(os.Args[2:])

	if *addr == "" {
		fmt.Println("Usage: balance --addr ADDR --node localhost:3000")
		return
	}

	chain, err := requestChain(*nodeAddr)
	if err != nil {
		fmt.Println("Balance request failed:", err)
		return
	}

	state := blockchain.NewState()
	for _, b := range chain {
		for _, tx := range b.Transactions {
			ok, _ := tx.VerifySignatureOnly()
			if !ok {
				fmt.Println("Chain contains invalid tx; cannot compute balance")
				return
			}
			_ = state.ApplyTx(tx)
		}
	}

	fmt.Printf("Balance(%s) = %d VLT\n", *addr, state.GetBalance(*addr))
}

// -------------------- networking helpers (P2P messages) --------------------

func sendTxToNode(addr string, tx blockchain.Transaction) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	raw, _ := json.Marshal(tx)
	msg := p2p.Message{Type: p2p.MsgTransaction, Data: raw}
	return json.NewEncoder(conn).Encode(msg)
}

func sendMineToNode(addr string, miner string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	payload, _ := json.Marshal(map[string]string{"miner": miner})
	msg := p2p.Message{Type: p2p.MsgMine, Data: payload}
	return json.NewEncoder(conn).Encode(msg)
}

func requestChain(addr string) ([]blockchain.Block, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)

	if err := enc.Encode(p2p.Message{Type: p2p.MsgGetChain, Data: []byte(`{}`)}); err != nil {
		return nil, err
	}

	var msg p2p.Message
	if err := dec.Decode(&msg); err != nil {
		return nil, err
	}
	if msg.Type != p2p.MsgChain {
		return nil, fmt.Errorf("unexpected response: %s", msg.Type)
	}

	var chain []blockchain.Block
	if err := json.Unmarshal(msg.Data, &chain); err != nil {
		return nil, err
	}
	return chain, nil
}

// -------------------- wallet helpers --------------------

// We avoid elliptic.Marshal in this file by encoding pubkey as fixed X||Y (P-256).
func walletFromPrivateKey(priv *ecdsa.PrivateKey) (*wallet.Wallet, error) {
	pub := publicKeyBytesFromPriv(priv)
	addr := wallet.AddressFromPublicKey(pub)

	return &wallet.Wallet{
		PrivateKey: priv,
		PublicKey:  pub,
		Address:    addr,
	}, nil
}

func publicKeyBytesFromPriv(priv *ecdsa.PrivateKey) []byte {
	x := priv.PublicKey.X.Bytes()
	y := priv.PublicKey.Y.Bytes()

	paddedX := make([]byte, 32)
	paddedY := make([]byte, 32)

	copy(paddedX[32-len(x):], x)
	copy(paddedY[32-len(y):], y)

	return append(paddedX, paddedY...)
}
